package mysql

import (
	"context"
	"crypto/rand"
	"database/sql"
	"os"
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
	return []models.UniqueConstraint{{Name: "unique_number", Fields: []string{"number"}}}
}
func (s physicalIndexed) GetIndexes() []models.Index {
	return []models.Index{{Name: "ordered_lookup", Fields: []string{"long", "id"}}}
}

func TestMySQLSchemaDeclarationNormalization(t *testing.T) {
	for _, tc := range []struct{ input, want string }{{"INT(11)", "int"}, {"tinyint(4)", "tinyint"}, {"int(11) unsigned", "int(11) unsigned"}, {"varchar(32)", "varchar(32)"}, {"DATETIME(3)", "datetime(3)"}} {
		if got := normalizeSchemaType(tc.input); got != tc.want {
			t.Fatalf("%+v: %s", tc, got)
		}
	}
	p := provider.NewLocalProvider("mysql-inspection", nil)
	model, err := p.RegisterModel(&physicalTypes{})
	if err != nil {
		t.Fatal(err)
	}
	builder := NewBuilder(p, codec.New(p, "tenant")).(*Builder)
	table, err := builder.DescribeTable(physicalIndexed{model})
	if err != nil {
		t.Fatal(err)
	}
	if table.Name != "tenant_PhysicalTypes" || table.Columns[0].Type != "bigint" || !table.Columns[0].AutoIncrement || table.Columns[0].Default != "" || !table.Columns[2].Nullable || table.Columns[3].Default != "0" || len(table.Indexes) != 3 {
		t.Fatal(table)
	}
}

// Requires a disposable MySQL 8.0.13+ database server. The random database is
// private to this test and is the only database the cleanup removes.
func TestMySQLPhysicalSchemaCatalog(t *testing.T) {
	dsn := os.Getenv("MAGICORM_SCHEMA_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set MAGICORM_SCHEMA_TEST_MYSQL_DSN to a disposable MySQL server")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := sql.Open("mysql", dsn)
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
	if _, err := conn.ExecContext(ctx, "CREATE DATABASE `"+schema+"`"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := conn.ExecContext(cleanupCtx, "DROP DATABASE `"+schema+"`"); err != nil {
			t.Error(err)
		}
	}()
	if _, err := conn.ExecContext(ctx, "USE `"+schema+"`"); err != nil {
		t.Fatal(err)
	}
	p := provider.NewLocalProvider("mysql-physical-inspect", nil)
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
	builder := NewBuilder(p, codec.New(p, "tenant")).(*Builder)
	expected, buildErr := builder.DescribeTable(model)
	if buildErr != nil {
		t.Fatal(buildErr)
	}
	executor := &ConnExecutor{dbConnPtr: conn}
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
	for _, tc := range []struct {
		name    string
		changes []string
	}{
		{"missing-column", []string{"ALTER TABLE `tenant_PhysicalTypes` DROP COLUMN `values`"}},
		{"wrong-type", []string{"ALTER TABLE `tenant_PhysicalTypes` MODIFY `name` varchar(30) NOT NULL"}},
		{"nullable", []string{"ALTER TABLE `tenant_PhysicalTypes` MODIFY `name` TEXT NULL"}},
		{"default", []string{"ALTER TABLE `tenant_PhysicalTypes` ALTER COLUMN `number` SET DEFAULT 9"}},
		{"no-auto", []string{"ALTER TABLE `tenant_PhysicalTypes` MODIFY `id` BIGINT NOT NULL"}},
		{"unsigned", []string{"ALTER TABLE `tenant_PhysicalTypes` MODIFY `number` INT UNSIGNED NOT NULL DEFAULT 0"}},
		{"wrong-order", []string{"DROP INDEX `ordered_lookup` ON `tenant_PhysicalTypes`", "CREATE INDEX `ordered_lookup` ON `tenant_PhysicalTypes` (`id`,`long`)"}},
		{"invisible-index", []string{"ALTER TABLE `tenant_PhysicalTypes` ALTER INDEX `ordered_lookup` INVISIBLE"}},
		{"descending-index", []string{"DROP INDEX `ordered_lookup` ON `tenant_PhysicalTypes`", "CREATE INDEX `ordered_lookup` ON `tenant_PhysicalTypes` (`long` DESC,`id`)"}},
		{"prefix-index", []string{"DROP INDEX `ordered_lookup` ON `tenant_PhysicalTypes`", "CREATE INDEX `ordered_lookup` ON `tenant_PhysicalTypes` (`optional`(10),`id`)"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, query := range tc.changes {
				execute(query)
			}
			inspect(false)
			execute("DROP TABLE `tenant_PhysicalTypes`")
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
	actual, readErr := executor.ReadTableSchema(ctx, relation.Name)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if issues := database.CompareTableSchema(relation, actual, true); len(issues) != 0 {
		t.Fatal(issues, relation, actual)
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
	execute("DROP TABLE `" + relation.Name + "`")
	actual, readErr = executor.ReadTableSchema(ctx, relation.Name)
	if readErr != nil || len(database.CompareTableSchema(relation, actual, false)) != 0 {
		t.Fatal(actual, readErr)
	}
}
