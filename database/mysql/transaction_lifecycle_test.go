package mysql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database"
	"testing"
)

type lifecycleConnector struct {
	begins, commits, rollbacks, statements int
	failBegin, failCommit, failRollback    bool
}
type lifecycleConnection struct{ owner *lifecycleConnector }
type lifecycleTransaction struct{ owner *lifecycleConnector }

func (c *lifecycleConnector) Connect(context.Context) (driver.Conn, error) {
	return &lifecycleConnection{c}, nil
}
func (*lifecycleConnector) Driver() driver.Driver { return serializableDriver{} }
func (*lifecycleConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected Prepare")
}
func (*lifecycleConnection) Close() error { return nil }
func (c *lifecycleConnection) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}
func (c *lifecycleConnection) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	c.owner.begins++
	if c.owner.failBegin {
		c.owner.failBegin = false
		return nil, errors.New("begin failure")
	}
	return &lifecycleTransaction{c.owner}, nil
}
func (c *lifecycleConnection) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	c.owner.statements++
	return driver.RowsAffected(1), nil
}
func (tx *lifecycleTransaction) Commit() error {
	tx.owner.commits++
	if tx.owner.failCommit {
		return errors.New("commit result unknown")
	}
	return nil
}
func (tx *lifecycleTransaction) Rollback() error {
	tx.owner.rollbacks++
	if tx.owner.failRollback {
		return errors.New("rollback failure")
	}
	return nil
}
func TestOrdinaryTransactionLifecycle(t *testing.T) {
	for _, kind := range []string{"host", "connection"} {
		for _, mode := range []string{"deferred", "begin-failure", "nested-success", "nested-rollback", "commit-failure", "rollback-failure", "released"} {
			t.Run(kind+"/"+mode, func(t *testing.T) {
				c := &lifecycleConnector{failBegin: mode == "begin-failure", failCommit: mode == "commit-failure", failRollback: mode == "rollback-failure"}
				db := sql.OpenDB(c)
				defer db.Close()
				var e database.Executor = &HostExecutor{dbHandle: db, executeContetxt: context.Background()}
				if kind == "connection" {
					conn, err := db.Conn(context.Background())
					if err != nil {
						t.Fatal(err)
					}
					e = &ConnExecutor{dbConnPtr: conn, executeContetxt: context.Background()}
				}
				defer e.Release()
				if mode == "released" {
					e.Release()
					if e.BeginTransaction() == nil || e.CommitTransaction() == nil || e.RollbackTransaction() == nil {
						t.Fatal("released executor accepted transaction")
					}
					return
				}
				if mode == "begin-failure" {
					if e.BeginTransaction() == nil {
						t.Fatal("missing begin failure")
					}
				}
				if err := e.BeginTransaction(); err != nil {
					t.Fatal(err)
				}
				switch mode {
				case "nested-success":
					if err := e.BeginTransaction(); err != nil {
						t.Fatal(err)
					}
					if err := e.CommitTransaction(); err != nil {
						t.Fatal(err)
					}
					if c.commits != 0 {
						t.Fatal("inner commit reached driver")
					}
					if _, err := e.Execute("statement"); err != nil {
						t.Fatal(err)
					}
				case "nested-rollback":
					if err := e.BeginTransaction(); err != nil {
						t.Fatal(err)
					}
					if err := e.RollbackTransaction(); err != nil {
						t.Fatal(err)
					}
					if e.BeginTransaction() == nil {
						t.Fatal("rollback-only Begin accepted")
					}
					if _, err := e.Query("statement", false); err == nil {
						t.Fatal("rollback-only Query accepted")
					}
					if _, err := e.Execute("statement"); err == nil {
						t.Fatal("rollback-only Execute accepted")
					}
					if e.CommitTransaction() == nil || c.commits != 0 || c.rollbacks != 1 || c.statements != 0 {
						t.Fatal("rollback-only transaction committed", c)
					}
				case "rollback-failure":
					if err := e.RollbackTransaction(); err == nil {
						t.Fatal("rollback failure lost")
					}
				default:
				}
				if mode != "nested-rollback" && mode != "rollback-failure" {
					err := e.CommitTransaction()
					if mode == "commit-failure" {
						if err == nil || err.Code != cd.Unexpected {
							t.Fatal("unknown commit changed", err)
						}
					} else if err != nil {
						t.Fatal(err)
					}
				}
				if err := e.RollbackTransaction(); err != nil {
					t.Fatal(err)
				}
				c.failCommit, c.failRollback = false, false
				before := c.begins
				if err := e.BeginTransaction(); err != nil {
					t.Fatal(err)
				}
				if c.begins != before+1 {
					t.Fatal("success without physical transaction", c)
				}
				if err := e.RollbackTransaction(); err != nil {
					t.Fatal(err)
				}
				if e.CommitTransaction() == nil {
					t.Fatal("idle commit accepted")
				}
			})
		}
	}
}

func TestTransactionThreeScopesAndCanceledBegin(t *testing.T) {
	for _, kind := range []string{"host", "connection"} {
		t.Run(kind, func(t *testing.T) {
			recorder := &lifecycleConnector{}
			db := sql.OpenDB(recorder)
			defer db.Close()
			canceled, cancel := context.WithCancel(context.Background())
			cancel()
			var executor database.Executor = &HostExecutor{dbHandle: db, executeContetxt: canceled}
			if kind == "connection" {
				conn, err := db.Conn(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				executor = &ConnExecutor{dbConnPtr: conn, executeContetxt: canceled}
			}
			defer executor.Release()
			if executor.BeginTransaction() == nil {
				t.Fatal("canceled begin admitted")
			}
			if err := executor.RollbackTransaction(); err != nil {
				t.Fatal(err)
			}
			switch value := executor.(type) {
			case *HostExecutor:
				value.executeContetxt = context.Background()
			case *ConnExecutor:
				value.executeContetxt = context.Background()
			}
			for range 3 {
				if err := executor.BeginTransaction(); err != nil {
					t.Fatal(err)
				}
			}
			if err := executor.RollbackTransaction(); err != nil {
				t.Fatal(err)
			}
			if err := executor.CommitTransaction(); err == nil {
				t.Fatal("middle rollback-only commit accepted")
			}
			if recorder.rollbacks != 0 || recorder.commits != 0 {
				t.Fatal("middle finalization ended physical transaction")
			}
			if _, err := executor.Execute("must reject"); err == nil {
				t.Fatal("rollback-only fell back to autocommit")
			}
			if err := executor.CommitTransaction(); err == nil || recorder.rollbacks != 1 || recorder.commits != 0 {
				t.Fatal("outer finalization did not roll back", err)
			}
			if err := executor.BeginTransaction(); err != nil {
				t.Fatal(err)
			}
			if err := executor.CommitTransaction(); err != nil {
				t.Fatal(err)
			}
			if recorder.begins != 2 || recorder.commits != 1 {
				t.Fatal("reuse did not establish actual new transaction", recorder)
			}
		})
	}
}
