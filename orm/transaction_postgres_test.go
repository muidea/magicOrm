//go:build !mysql

package orm

import (
	"context"
	"database/sql"
	"encoding/binary"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/internal/testdatabase"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

type commitReceipt struct {
	ID    int64  `orm:"id key" view:"detail,lite"`
	Value string `orm:"value" view:"detail,lite"`
}

// PostgreSQL wire proxy: relay COMMIT to the real database, consume the server's
// success/ready messages, then disconnect without delivering the acknowledgement.
// No fake driver outcome and no replay: a separate direct connection proves fact.
func commitResponseLossProxy(t *testing.T, upstream string, forwardCommit bool) (string, <-chan bool) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	completed := make(chan bool, 1)
	t.Cleanup(func() { _ = listener.Close() })
	go func() {
		client, err := listener.Accept()
		if err != nil {
			return
		}
		defer client.Close()
		server, err := net.DialTimeout("tcp", upstream, 5*time.Second)
		if err != nil {
			return
		}
		defer server.Close()
		_ = client.SetDeadline(time.Now().Add(20 * time.Second))
		_ = server.SetDeadline(time.Now().Add(20 * time.Second))
		var dropping atomic.Bool
		go func() {
			defer server.Close()
			header := make([]byte, 4)
			if _, err := io.ReadFull(client, header); err != nil {
				return
			}
			size := int(binary.BigEndian.Uint32(header))
			if size < 8 || size > 1<<20 {
				return
			}
			body := make([]byte, size-4)
			if _, err := io.ReadFull(client, body); err != nil {
				return
			}
			if _, err := server.Write(append(header, body...)); err != nil {
				return
			}
			for {
				header := make([]byte, 5)
				if _, err := io.ReadFull(client, header); err != nil {
					return
				}
				size := int(binary.BigEndian.Uint32(header[1:]))
				if size < 4 || size > 1<<20 {
					return
				}
				body := make([]byte, size-4)
				if _, err := io.ReadFull(client, body); err != nil {
					return
				}
				if header[0] == 'Q' && strings.EqualFold(strings.TrimRight(string(body), "\x00"), "COMMIT") {
					dropping.Store(true)
					if !forwardCommit {
						completed <- false
						_ = client.Close()
						return
					}
				}
				if _, err := server.Write(append(header, body...)); err != nil {
					return
				}
			}
		}()
		commitOK := false
		for {
			header := make([]byte, 5)
			if _, err := io.ReadFull(server, header); err != nil {
				return
			}
			size := int(binary.BigEndian.Uint32(header[1:]))
			if size < 4 || size > 1<<20 {
				return
			}
			body := make([]byte, size-4)
			if _, err := io.ReadFull(server, body); err != nil {
				return
			}
			if dropping.Load() {
				if header[0] == 'C' && string(body) == "COMMIT\x00" {
					commitOK = true
				}
				if header[0] == 'Z' {
					completed <- commitOK
					return
				}
				continue
			}
			if _, err := client.Write(append(header, body...)); err != nil {
				return
			}
		}
	}()
	return listener.Addr().String(), completed
}

func TestPostgresUnknownCommitMustBeReconciled(t *testing.T) {
	dsn := os.Getenv("MAGICORM_SCHEMA_TEST_POSTGRES_DSN")
	u := testdatabase.RequirePostgres(t, dsn)
	for _, committed := range []bool{true, false} {
		t.Run(strconv.FormatBool(committed), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			db, err := sql.Open("postgres", dsn)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			schema := "commit_receipt_" + strconv.FormatBool(committed)
			if _, err := db.ExecContext(ctx, `CREATE SCHEMA "`+schema+`"`); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if _, err := db.Exec(`DROP SCHEMA "` + schema + `" CASCADE`); err != nil {
					t.Error(err)
				}
			}()
			password, _ := u.User.Password()
			p := provider.NewLocalProvider("unknown-commit", nil)
			model, ce := p.RegisterModel(&commitReceipt{})
			if ce != nil {
				t.Fatal(ce)
			}
			direct, ce := NewOrm(p, NewConfig(u.Host, strings.TrimPrefix(u.Path, "/"), schema, u.User.Username(), password), "receipt")
			if ce != nil {
				t.Fatal(ce)
			}
			defer direct.Release()
			if ce := direct.Create(model); ce != nil {
				t.Fatal(ce)
			}
			endpoint, completed := commitResponseLossProxy(t, u.Host, committed)
			uncertain, ce := NewOrm(p, NewConfig(endpoint, strings.TrimPrefix(u.Path, "/"), schema, u.User.Username(), password), "receipt")
			if ce != nil {
				t.Fatal(ce)
			}
			defer uncertain.Release()
			if ce := uncertain.BeginTransaction(); ce != nil {
				t.Fatal(ce)
			}
			value, ce := p.GetEntityModel(&commitReceipt{ID: 712, Value: "original-operation"}, true)
			if ce != nil {
				t.Fatal(ce)
			}
			if _, ce := uncertain.Insert(value); ce != nil {
				t.Fatal(ce)
			}
			if ce := uncertain.CommitTransaction(); ce == nil || ce.Code != cd.Unexpected {
				t.Fatalf("lost commit acknowledgement must stay unknown: %v", ce)
			}
			select {
			case ok := <-completed:
				if ok != committed {
					t.Fatal("proxy did not establish expected server outcome")
				}
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
			query, ce := p.GetEntityModel(&commitReceipt{ID: 712}, true)
			if ce != nil {
				t.Fatal(ce)
			}
			receipt, ce := direct.Query(query)
			if committed {
				if ce != nil || receipt.Interface(true).(*commitReceipt).Value != "original-operation" {
					t.Fatal("independent committed identity lookup failed", ce)
				}
			} else if ce == nil || ce.Code != cd.NotFound {
				t.Fatal("unsent commit unexpectedly persisted", ce)
			}
			filter, ce := p.GetEntityFilter(&commitReceipt{}, models.MetaView)
			if ce != nil {
				t.Fatal(ce)
			}
			count, ce := direct.Count(filter)
			expected := int64(0)
			if committed {
				expected = 1
			}
			if ce != nil || count != expected {
				t.Fatal("unknown commit must not trigger automatic replay", count, ce)
			}
		})
	}
}
