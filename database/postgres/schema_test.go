package postgres

import (
	"context"
	"crypto/rand"
	"database/sql"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

type physicalTypes struct {
	ID       int64          `orm:"id key auto"`
	Name     string         `orm:"name"`
	Optional *string        `orm:"optional"`
	Flag     bool           `orm:"flag"`
	Byte     int8           `orm:"byte"`
	Small    int16          `orm:"small"`
	Number   int            `orm:"number"`
	Long     int64          `orm:"long"`
	Unsigned uint64         `orm:"unsigned"`
	Real     float32        `orm:"real"`
	Double   float64        `orm:"double"`
	Time     time.Time      `orm:"time"`
	Values   []int          `orm:"values"`
	Related  []*physicalKey `orm:"related"`
}
type physicalKey struct {
	ID string `orm:"id key"`
}
type physicalIndexed struct{ models.Model }

func (s physicalIndexed) GetUniqueConstraints() []models.UniqueConstraint {
	return []models.UniqueConstraint{{Name: "unique_name", Fields: []string{"name"}}}
}
func (s physicalIndexed) GetIndexes() []models.Index {
	return []models.Index{{Name: "ordered_lookup", Fields: []string{"optional", "id"}}}
}

func TestPostgresSchemaDeclarationNormalization(t *testing.T) {
	for _, tc := range []struct{ kind, input, want string }{{"bigint", "'0'::bigint", "0"}, {"integer", "0", "0"}, {"boolean", "'1'", "true"}, {"boolean", "false", "false"}, {"double precision", "'1.5'::double precision", "1.5"}, {"bigint", "'0'::text", "'0'::text"}} {
		if got := normalizeSchemaDefault(tc.input, tc.kind); got != tc.want {
			t.Fatalf("%+v: %s", tc, got)
		}
	}
}

// Opt-in only. The DSN must point at a disposable test database whose owner can
// create schemas. Each test owns a random schema and removes only that schema.
func TestPostgresPhysicalSchemaCatalog(t *testing.T) {
	dsn := os.Getenv("MAGICORM_SCHEMA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set MAGICORM_SCHEMA_TEST_POSTGRES_DSN to a disposable PostgreSQL database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	schema := "orm_inspect_" + strings.ToLower(rand.Text())
	if _, err := conn.ExecContext(ctx, `CREATE SCHEMA "`+schema+`"`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := conn.ExecContext(cleanupCtx, `DROP SCHEMA "`+schema+`" CASCADE`); err != nil {
			t.Error(err)
		}
	}()
	if _, err := conn.ExecContext(ctx, `SET search_path TO "`+schema+`"`); err != nil {
		t.Fatal(err)
	}
	p := provider.NewLocalProvider("physical-inspect", nil)
	for _, value := range []any{&physicalTypes{}, &physicalKey{}} {
		if _, err := p.RegisterModel(value); err != nil {
			t.Fatal(err)
		}
	}
	model, modelErr := p.GetEntityModel(&physicalTypes{}, true)
	if modelErr != nil {
		t.Fatal(modelErr)
	}
	model = physicalIndexed{model}
	builder := NewBuilder(p, codec.New(p, "tenant"))
	expected, buildErr := builder.DescribeTable(model)
	if buildErr != nil {
		t.Fatal(buildErr)
	}
	executor := &ConnExecutor{dbConnPtr: conn, schemaName: schema}
	execute := func(query string) {
		t.Helper()
		if _, err := conn.ExecContext(ctx, query); err != nil {
			t.Fatal(err)
		}
	}
	create := func() {
		t.Helper()
		result, err := builder.BuildCreateTable(model)
		if err != nil {
			t.Fatal(err)
		}
		execute(result.SQL())
	}
	inspect := func(wantMatch bool) {
		t.Helper()
		actual, err := executor.ReadTableSchema(ctx, expected.Name)
		if err != nil {
			t.Fatal(err)
		}
		issues := database.CompareTableSchema(expected, actual, true)
		if (len(issues) == 0) != wantMatch {
			t.Fatalf("issues=%+v\nexpected=%+v\nactual=%+v", issues, expected, actual)
		}
	}
	create()
	inspect(true)
	// A second schema with the same table name must not satisfy a missing table.
	foreign := &ConnExecutor{dbConnPtr: conn, schemaName: schema + "_missing"}
	actual, readErr := foreign.ReadTableSchema(ctx, expected.Name)
	if readErr != nil || actual.Exists {
		t.Fatal(actual, readErr)
	}
	// Quoted input remains a bound value, never executable SQL.
	actual, readErr = executor.ReadTableSchema(ctx, `x'; DROP SCHEMA public CASCADE; --`)
	if readErr != nil || actual.Exists {
		t.Fatal(actual, readErr)
	}
	for _, tc := range []struct{ name, change string }{
		{"missing-column", `ALTER TABLE "tenant_PhysicalTypes" DROP COLUMN "values"`},
		{"wrong-type", `ALTER TABLE "tenant_PhysicalTypes" ALTER COLUMN "name" TYPE varchar(30)`},
		{"nullable", `ALTER TABLE "tenant_PhysicalTypes" ALTER COLUMN "name" DROP NOT NULL`},
		{"default", `ALTER TABLE "tenant_PhysicalTypes" ALTER COLUMN "number" SET DEFAULT 9`},
		{"missing-serial", `ALTER TABLE "tenant_PhysicalTypes" ALTER COLUMN "id" DROP DEFAULT`},
		{"wrong-order", `DROP INDEX "ordered_lookup"; CREATE INDEX "ordered_lookup" ON "tenant_PhysicalTypes" ("id","optional")`},
		{"partial-index", `DROP INDEX "ordered_lookup"; CREATE INDEX "ordered_lookup" ON "tenant_PhysicalTypes" ("optional","id") WHERE "id" > 0`},
		{"include-index", `DROP INDEX "ordered_lookup"; CREATE INDEX "ordered_lookup" ON "tenant_PhysicalTypes" ("optional","id") INCLUDE ("name")`},
		{"expression-index", `DROP INDEX "ordered_lookup"; CREATE INDEX "ordered_lookup" ON "tenant_PhysicalTypes" (lower("optional"),"id")`},
		{"descending-index", `DROP INDEX "ordered_lookup"; CREATE INDEX "ordered_lookup" ON "tenant_PhysicalTypes" ("optional" DESC,"id")`},
		{"unique-drift", `ALTER TABLE "tenant_PhysicalTypes" DROP CONSTRAINT "unique_name"; ALTER TABLE "tenant_PhysicalTypes" ADD CONSTRAINT "unique_name" UNIQUE ("optional")`},
		{"view", `DROP TABLE "tenant_PhysicalTypes"; CREATE VIEW "tenant_PhysicalTypes" AS SELECT 1 AS id`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			execute(tc.change)
			inspect(false)
			if tc.name == "view" {
				execute(`DROP VIEW "tenant_PhysicalTypes"`)
			} else {
				execute(`DROP TABLE "tenant_PhysicalTypes"`)
			}
			create()
			inspect(true)
		})
	}
	field := model.GetField("related")
	relation, buildErr := builder.DescribeRelationTable(model, field)
	if buildErr != nil {
		t.Fatal(buildErr)
	}
	result, buildErr := builder.BuildCreateRelationTable(model, field)
	if buildErr != nil {
		t.Fatal(buildErr)
	}
	execute(result.SQL())
	actual, readErr = executor.ReadTableSchema(ctx, relation.Name)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if issues := database.CompareTableSchema(relation, actual, true); len(issues) != 0 {
		t.Fatal(issues, relation, actual)
	}
	// The host executor and transactional connection paths must read the same scope.
	host := &HostExecutor{dbHandle: db, schemaName: schema}
	hostActual, readErr := host.ReadTableSchema(ctx, relation.Name)
	if readErr != nil || !reflect.DeepEqual(actual, hostActual) {
		t.Fatal(readErr, actual, hostActual)
	}
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	executor.dbTx = tx
	_, readErr = executor.ReadTableSchema(ctx, relation.Name)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	executor.dbTx = nil
	execute(`DROP TABLE "` + relation.Name + `"`)
	actual, readErr = executor.ReadTableSchema(ctx, relation.Name)
	if readErr != nil || len(database.CompareTableSchema(relation, actual, false)) != 0 {
		t.Fatal(actual, readErr)
	}
}
