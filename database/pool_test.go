package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type poolTestConnector struct{ opened atomic.Int64 }

func (c *poolTestConnector) Connect(context.Context) (driver.Conn, error) {
	c.opened.Add(1)
	return poolTestConn{}, nil
}
func (c *poolTestConnector) Driver() driver.Driver { return poolTestDriver{} }

type poolTestDriver struct{}

func (poolTestDriver) Open(string) (driver.Conn, error) { return poolTestConn{}, nil }

type poolTestConn struct{}

func (poolTestConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("unused") }
func (poolTestConn) Close() error                        { return nil }
func (poolTestConn) Begin() (driver.Tx, error)           { return nil, errors.New("unused") }

func TestPoolReusesConcurrentBurstsWithinConnectionBudget(t *testing.T) {
	for _, limit := range []int{1, 2, 8} {
		connector := &poolTestConnector{}
		db := sql.OpenDB(connector)
		t.Cleanup(func() { _ = db.Close() })
		ConfigurePool(db, limit)
		for burst := 0; burst < 3; burst++ {
			var connections []*sql.Conn
			for i := 0; i < limit; i++ {
				conn, err := db.Conn(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				connections = append(connections, conn)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			conn, err := db.Conn(ctx)
			cancel()
			if conn != nil || !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("connection budget %d was exceeded: %v", limit, err)
			}
			for _, conn := range connections {
				if err := conn.Close(); err != nil {
					t.Fatal(err)
				}
			}
		}
		stats := db.Stats()
		if connector.opened.Load() != int64(limit) || stats.Idle != limit || stats.MaxIdleClosed != 0 {
			t.Fatalf("limit=%d reopened connections across bursts: opened=%d stats=%+v", limit, connector.opened.Load(), stats)
		}
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestUnboundedPoolRetainsBoundedIdleConnections(t *testing.T) {
	for _, limit := range []int{0, -1} {
		db := sql.OpenDB(&poolTestConnector{})
		defer db.Close()
		ConfigurePool(db, limit)
		var connections []*sql.Conn
		for i := 0; i < 8; i++ {
			conn, err := db.Conn(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			connections = append(connections, conn)
		}
		for _, conn := range connections {
			_ = conn.Close()
		}
		if stats := db.Stats(); stats.Idle != 2 || stats.MaxOpenConnections != 0 {
			t.Fatalf("unbounded pool idle policy changed: %+v", stats)
		}
	}
}

func TestAcquireConnectionPreservesBudgetAndCancellation(t *testing.T) {
	db := sql.OpenDB(&poolTestConnector{})
	defer db.Close()
	ConfigurePool(db, 1)
	held, err := AcquireConnection(context.Background(), db, DatabasePostgreSQL)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	blocked, err := AcquireConnection(ctx, db, DatabasePostgreSQL)
	if blocked != nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("acquisition ignored pool budget/cancellation: conn=%v err=%v", blocked, err)
	}
	if err := held.Close(); err != nil {
		t.Fatal(err)
	}
	reused, err := AcquireConnection(context.Background(), db, DatabaseMySQL)
	if err != nil {
		t.Fatal(err)
	}
	defer reused.Close()
	if db.Stats().OpenConnections != 1 {
		t.Fatal("acquisition opened excess connections")
	}
}
