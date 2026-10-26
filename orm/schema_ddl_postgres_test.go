//go:build !mysql

package orm

import (
	"context"
	"crypto/rand"
	"database/sql"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/muidea/magicOrm/internal/testdatabase"

	"github.com/muidea/magicOrm/internal/database"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
)

// Opt-in isolated database test. It owns a random schema, never reuses public
// or changes objects outside that schema. The environment value must be a
// PostgreSQL URL pointing to a disposable test database.
func TestPostgresSchemaDDLPreflightAndOwnedReconcile(t *testing.T) {
	dsn := os.Getenv("MAGICORM_SCHEMA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set MAGICORM_SCHEMA_TEST_POSTGRES_DSN to a disposable PostgreSQL URL")
	}
	testdatabase.RequirePostgres(t, dsn)
	u, err := url.Parse(dsn)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.User == nil || u.Host == "" || u.Path == "" {
		t.Fatal("test database must use a PostgreSQL URL with explicit credentials")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	schema := "ddl_preflight_" + strings.ToLower(rand.Text())
	if _, err := db.ExecContext(ctx, `CREATE SCHEMA "`+schema+`"`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanup, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		if _, err := db.ExecContext(cleanup, `DROP SCHEMA "`+schema+`" CASCADE`); err != nil {
			t.Error(err)
		}
	}()
	password, _ := u.User.Password()
	cfg := newConfig(u.Host, strings.TrimPrefix(u.Path, "/"), schema, u.User.Username(), password)
	open := func(p provider.Provider, prefix string) *impl {
		t.Helper()
		handler, err := NewOrm(p, cfg, prefix)
		if err != nil {
			t.Fatal(err)
		}
		return handler.(*impl)
	}
	t.Run("index-storage-admission-and-physical-index", func(t *testing.T) {
		for _, inline := range []bool{false, true} {
			p := provider.NewRemoteProvider("index-storage", nil)
			target, err := helper.GetObject(&inspectReferenced{})
			if err != nil {
				t.Fatal(err)
			}
			if _, err = p.RegisterModel(target); err != nil {
				t.Fatal(err)
			}
			object, err := helper.GetObject(&indexStorageRecord{})
			if err != nil {
				t.Fatal(err)
			}
			object.Indexes = []models.Index{{Name: "idx_reference", Fields: []string{"related", "name"}}}
			if inline {
				object.UniqueConstraints = []models.UniqueConstraint{{Name: "uq_reference", Fields: []string{"related", "name"}}}
			}
			model, err := p.RegisterModel(object)
			if err != nil {
				t.Fatal(err)
			}
			prefix := "invalid_index"
			if inline {
				prefix = "valid_index"
			}
			handler := open(p, prefix)
			defer handler.Release()
			err = handler.Create(model)
			if !inline {
				if err == nil {
					t.Fatal("relation-table index accepted")
				}
				if report, e := handler.InspectSchema(model, false); e != nil || !report.Matches || len(report.Tables) != 2 {
					t.Fatal("invalid DDL left physical objects or blocked diagnosis", report, e)
				}
				if _, e := handler.RepairSchema(nil, model); e == nil {
					t.Fatal("repair accepted invalid index")
				}
				continue
			}
			if err != nil {
				t.Fatal(err)
			}
			if report, e := handler.InspectSchema(model, true); e != nil || !report.Matches {
				t.Fatal("valid inline index not physically created", report, e)
			}
		}
	})
	t.Run("create-drop-owned-and-reference-boundary", func(t *testing.T) {
		fixture, model, _ := inspectionFixture(t)
		handler := open(fixture.modelProvider, "graph")
		defer handler.Release()
		if err := handler.Create(model); err != nil {
			t.Fatal(err)
		}
		report, err := handler.InspectSchema(model, true)
		if err != nil || !report.Matches || len(report.Tables) != 4 {
			t.Fatal(report, err)
		}
		if err := handler.Drop(model); err != nil {
			t.Fatal(err)
		}
		report, err = handler.InspectSchema(model, false)
		if err != nil || !report.Matches {
			t.Fatal(report, err)
		}
	})
	t.Run("add-owned-component", func(t *testing.T) {
		p, previous, current := reconcileModels(t, &reconcileNewOwned{})
		object, err := helper.GetObject(&inspectOwned{})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.RegisterModel(object); err != nil {
			t.Fatal(err)
		}
		handler := open(p, "evolve")
		defer handler.Release()
		if err := handler.Create(previous); err != nil {
			t.Fatal(err)
		}
		if err := handler.Reconcile(previous, current); err != nil {
			t.Fatal(err)
		}
		report, err := handler.InspectSchema(current, true)
		if err != nil || !report.Matches || len(report.Tables) != 3 {
			t.Fatal(report, err)
		}
	})
	t.Run("late-rejection-leaves-old-schema-exact", func(t *testing.T) {
		p, previous, current := reconcileModels(t, &reconcileLateRequired{})
		handler := open(p, "rejected")
		defer handler.Release()
		if err := handler.Create(previous); err != nil {
			t.Fatal(err)
		}
		if err := handler.Reconcile(previous, current); err == nil {
			t.Fatal("required addition was accepted")
		}
		report, err := handler.InspectSchema(previous, true)
		if err != nil || !report.Matches {
			t.Fatal("earlier nullable addition was executed", report, err)
		}
	})
	t.Run("repair-partial-create-and-drop", func(t *testing.T) {
		fixture, model, _ := inspectionFixture(t)
		handler := open(fixture.modelProvider, "repair_graph")
		defer handler.Release()
		builder := newBuilder(handler.modelProvider, handler.modelCodec)
		statement, err := builder.BuildCreateTable(model)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := handler.executor.Execute(statement.SQL(), statement.Args()...); err != nil {
			t.Fatal(err)
		}
		for range 2 {
			report, err := handler.RepairSchema(nil, model)
			if err != nil || report == nil || !report.Matches || len(report.Tables) != 4 {
				t.Fatal(report, err)
			}
		}
		statement, err = builder.BuildDropTable(model)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := handler.executor.Execute(statement.SQL(), statement.Args()...); err != nil {
			t.Fatal(err)
		}
		for range 2 {
			report, err := handler.RepairSchema(model, nil)
			if err != nil || report == nil || !report.Matches {
				t.Fatal(report, err)
			}
		}
	})
	t.Run("repair-additive-keeps-existing-data", func(t *testing.T) {
		p, previous, current := reconcileModels(t, &reconcileCurrentModel{})
		current = indexedRepairModel(current)
		handler := open(p, "repair_add")
		defer handler.Release()
		if err := handler.Create(previous); err != nil {
			t.Fatal(err)
		}
		name := handler.modelCodec.ConstructModelTableName(previous)
		if _, err := handler.executor.Execute(`INSERT INTO "` + name + `" ("id", "namespace") VALUES (1, 'preserve-me')`); err != nil {
			t.Fatal(err)
		}
		for range 2 {
			report, err := handler.RepairSchema(previous, current)
			if err != nil || report == nil || !report.Matches {
				t.Fatal(report, err)
			}
		}
		var value string
		if err := db.QueryRowContext(ctx, `SELECT "namespace" FROM "`+schema+`"."`+name+`" WHERE "id"=1`).Scan(&value); err != nil || value != "preserve-me" {
			t.Fatal(value, err)
		}
	})
	t.Run("repair-rejects-drift-before-creating-missing-table", func(t *testing.T) {
		fixture, model, _ := inspectionFixture(t)
		handler := open(fixture.modelProvider, "repair_drift")
		defer handler.Release()
		owned, err := handler.modelProvider.GetTypeModel(model.GetField("owned").GetType().Elem())
		if err != nil {
			t.Fatal(err)
		}
		if err := handler.Create(owned); err != nil {
			t.Fatal(err)
		}
		name := handler.modelCodec.ConstructModelTableName(owned)
		if _, err := handler.executor.Execute(`ALTER TABLE "` + name + `" ADD COLUMN "foreign_column" TEXT`); err != nil {
			t.Fatal(err)
		}
		if report, err := handler.RepairSchema(nil, model); err == nil || report != nil {
			t.Fatal(report, err)
		}
		host, err := handler.executor.(database.SchemaReader).ReadTableSchema(ctx, handler.modelCodec.ConstructModelTableName(model))
		if err != nil || host.Exists {
			t.Fatal("rejected repair partially created host", host, err)
		}
	})
}

type indexStorageRecord struct {
	ID      int64              `orm:"id key"`
	Name    string             `orm:"name"`
	Related *inspectReferenced `orm:"related"`
}
