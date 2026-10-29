package orm

import (
	"slices"
	"strings"
	"testing"

	"github.com/muidea/magicOrm/internal/database"
	"github.com/muidea/magicOrm/models"
)

type inspectSecondHost struct {
	ID    int64        `orm:"id key auto"`
	Owned inspectOwned `orm:"owned"`
}

func TestCompleteSchemaDropDeduplicatesSharedOwnedAndResumes(t *testing.T) {
	h, host, catalog := inspectionFixture(t)
	owned, err := h.modelProvider.GetEntityModel(&inspectOwned{}, true)
	if err != nil {
		t.Fatal(err)
	}
	second, err := h.modelProvider.RegisterModel(&inspectSecondHost{})
	if err != nil {
		t.Fatal(err)
	}
	graph, err := collectSchemaGraph(h.context, second, h.modelProvider, h.modelCodec)
	if err != nil {
		t.Fatal(err)
	}
	for name, table := range graph.tables {
		catalog.tables[name] = table
	}
	roots := []models.Model{host, owned, second}
	tables, err := SchemaSetTables(h.context, roots, h.modelProvider, "tenant_a")
	if err != nil || len(tables) != 6 {
		t.Fatal(tables, err)
	}
	executor := &repairExecutor{schemaExecutor: catalog, fail: true}
	executor.apply = func(sql string) {
		for name := range catalog.tables {
			if strings.Contains(sql, dialectSQLForTest(`"`+name+`"`)) {
				delete(catalog.tables, name)
			}
		}
	}
	h.executor = executor
	if _, err := h.DropSchemaSet(roots); err == nil || len(executor.execCalls) != 1 {
		t.Fatal("missing ambiguous failure", err)
	}
	executor.fail = false
	report, err := h.DropSchemaSet(roots)
	if err != nil || report == nil || !report.Matches || !slices.Equal(report.Tables, tables) || len(executor.execCalls) != 6 {
		t.Fatal(report, err, executor.execCalls)
	}
	if _, err := h.DropSchemaSet(roots); err != nil || len(executor.execCalls) != 6 {
		t.Fatal("replayed completed DDL", err)
	}
	if _, err := h.DropSchemaSet([]models.Model{host, host}); err == nil {
		t.Fatal("duplicate roots admitted")
	}
}

func TestCompleteSchemaDropChecksEntireGraphBeforeFirstWrite(t *testing.T) {
	h, host, catalog := inspectionFixture(t)
	for name, table := range catalog.tables {
		table.Columns = append(table.Columns, database.ColumnSchema{Name: "foreign"})
		catalog.tables[name] = table
		break
	}
	if _, err := h.DropSchemaSet([]models.Model{host}); err == nil || len(catalog.execCalls) != 0 {
		t.Fatal("drift changed storage", err, catalog.execCalls)
	}
}
