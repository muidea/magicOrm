package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/muidea/magicOrm/internal/database"
)

type serializableConnector struct {
	options []driver.TxOptions
	fail    bool
}

type serializableDriver struct{}
type serializableConn struct{ owner *serializableConnector }
type serializableTx struct{}

func (s *serializableConnector) Connect(context.Context) (driver.Conn, error) {
	return &serializableConn{owner: s}, nil
}
func (*serializableConnector) Driver() driver.Driver { return serializableDriver{} }
func (serializableDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("unexpected Open")
}
func (*serializableConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected Prepare")
}
func (*serializableConn) Close() error { return nil }
func (*serializableConn) Begin() (driver.Tx, error) {
	return nil, errors.New("unexpected legacy Begin")
}
func (s *serializableConn) BeginTx(_ context.Context, options driver.TxOptions) (driver.Tx, error) {
	s.owner.options = append(s.owner.options, options)
	if s.owner.fail {
		s.owner.fail = false
		return nil, errors.New("begin rejected")
	}
	return serializableTx{}, nil
}
func (serializableTx) Commit() error   { return nil }
func (serializableTx) Rollback() error { return nil }

// 内存 driver 只核对隔离选项与事务边界，不作为数据库并发证据。
func TestSerializableTransactionOptions(t *testing.T) {
	for _, kind := range []string{"host", "connection"} {
		for _, mode := range []string{"normal", "begin-failure", "cancel"} {
			t.Run(kind+"/"+mode, func(t *testing.T) {
				connector := &serializableConnector{fail: mode == "begin-failure"}
				db := sql.OpenDB(connector)
				defer db.Close()
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				var executor database.Executor = &HostExecutor{executeContetxt: ctx, dbHandle: db}
				if kind == "connection" {
					conn, err := db.Conn(ctx)
					if err != nil {
						t.Fatal(err)
					}
					executor = &ConnExecutor{executeContetxt: ctx, dbConnPtr: conn}
				}
				defer executor.Release()
				capability := executor.(database.SerializableTransactionExecutor)
				if mode == "cancel" {
					cancel()
					if capability.BeginSerializableTransaction() == nil {
						t.Fatal("取消后仍开启事务")
					}
					return
				}
				if mode == "begin-failure" && capability.BeginSerializableTransaction() == nil {
					t.Fatal("开启失败被当作成功")
				}
				if err := capability.BeginSerializableTransaction(); err != nil {
					t.Fatal(err)
				}
				if capability.BeginSerializableTransaction() == nil {
					t.Fatal("已有事务允许伪升级")
				}
				if err := executor.BeginTransaction(); err != nil {
					t.Fatal(err)
				}
				if err := executor.CommitTransaction(); err != nil {
					t.Fatal(err)
				}
				if err := executor.CommitTransaction(); err != nil {
					t.Fatal(err)
				}
				if err := capability.BeginSerializableTransaction(); err != nil {
					t.Fatal(err)
				}
				if err := executor.RollbackTransaction(); err != nil {
					t.Fatal(err)
				}
				for _, options := range connector.options {
					if options.Isolation != driver.IsolationLevel(sql.LevelSerializable) || options.ReadOnly {
						t.Fatal("事务隔离选项错误", options)
					}
				}
			})
		}
	}
}
