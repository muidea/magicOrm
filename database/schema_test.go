package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
)

func TestComparePhysicalSchema(t *testing.T) {
	baseline := func() TableSchema {
		return TableSchema{Name: "ns_Entity", Exists: true, Ordinary: true, Columns: []ColumnSchema{{Name: "id", Type: "bigint", AutoIncrement: true}, {Name: "value", Type: "text", Nullable: true}}, Indexes: []IndexSchema{{Name: "PRIMARY", Columns: []string{"id"}, Primary: true, Unique: true, Constraint: true, Valid: true}, {Name: "lookup", Columns: []string{"value", "id"}, Valid: true}}}
	}
	for _, tc := range []struct {
		name   string
		change func(*TableSchema)
		kind   string
	}{
		{"match", func(*TableSchema) {}, ""},
		{"missing-table", func(v *TableSchema) { v.Exists = false }, "table-missing"},
		{"view", func(v *TableSchema) { v.Ordinary = false }, "table-kind-mismatch"},
		{"missing-column", func(v *TableSchema) { v.Columns = v.Columns[:1] }, "column-missing"},
		{"type", func(v *TableSchema) { v.Columns[1].Type = "integer" }, "column-mismatch"},
		{"nullable", func(v *TableSchema) { v.Columns[1].Nullable = false }, "column-mismatch"},
		{"default", func(v *TableSchema) { v.Columns[0].Default = "9" }, "column-mismatch"},
		{"auto", func(v *TableSchema) { v.Columns[0].AutoIncrement = false }, "column-mismatch"},
		{"generated", func(v *TableSchema) { v.Columns[1].Generated = true }, "column-mismatch"},
		{"extra-column", func(v *TableSchema) { v.Columns = append(v.Columns, ColumnSchema{Name: "extra"}) }, "column-unexpected"},
		{"duplicate-column", func(v *TableSchema) { v.Columns = append(v.Columns, v.Columns[0]) }, "column-duplicate"},
		{"primary", func(v *TableSchema) { v.Indexes = v.Indexes[1:] }, "index-missing"},
		{"order", func(v *TableSchema) { v.Indexes[1].Columns = []string{"id", "value"} }, "index-mismatch"},
		{"unique", func(v *TableSchema) { v.Indexes[1].Unique = true }, "index-mismatch"},
		{"constraint", func(v *TableSchema) { v.Indexes[0].Constraint = false }, "index-mismatch"},
		{"invalid-index", func(v *TableSchema) { v.Indexes[1].Valid = false }, "index-mismatch"},
		{"extra-index", func(v *TableSchema) { v.Indexes = append(v.Indexes, IndexSchema{Name: "extra"}) }, "index-unexpected"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			expected, actual := baseline(), baseline()
			tc.change(&actual)
			issues := CompareTableSchema(expected, actual, true)
			if tc.kind == "" {
				if len(issues) != 0 {
					t.Fatal(issues)
				}
				return
			}
			if len(issues) != 1 || issues[0].Kind != tc.kind {
				t.Fatal(issues)
			}
		})
	}
	if issues := CompareTableSchema(baseline(), baseline(), false); len(issues) != 1 || issues[0].Kind != "table-still-exists" {
		t.Fatal(issues)
	}
	if issues := CompareTableSchema(baseline(), TableSchema{}, false); len(issues) != 0 {
		t.Fatal(issues)
	}
}

func TestNormalizeSchemaLiteral(t *testing.T) {
	for _, tc := range []struct{ kind, input, want string }{{"bigint", "'0'", "0"}, {"integer", "'-12'", "-12"}, {"integer", "1+1", "1+1"}, {"text", "'1'", "'1'"}, {"double precision", "'1.50'", "1.5"}, {"boolean", "'0'", "0"}, {"bigint", "nextval('x')", "nextval('x')"}} {
		if got := NormalizeSchemaLiteral(tc.input, tc.kind); got != tc.want {
			t.Fatalf("%+v: %q", tc, got)
		}
	}
}

type schemaTestDriver struct{}
type schemaTestConn struct{ mode string }
type schemaTestRows struct {
	mode string
	read bool
}

func (schemaTestDriver) Open(name string) (driver.Conn, error) {
	return &schemaTestConn{mode: name}, nil
}
func (*schemaTestConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}
func (*schemaTestConn) Close() error              { return nil }
func (*schemaTestConn) Begin() (driver.Tx, error) { return nil, errors.New("unexpected transaction") }
func (s *schemaTestConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	if s.mode == "query-error" {
		return nil, errors.New("query failed")
	}
	return &schemaTestRows{mode: s.mode}, nil
}
func (*schemaTestRows) Columns() []string { return []string{"value"} }
func (s *schemaTestRows) Close() error {
	if s.mode == "close-error" {
		return errors.New("close failed")
	}
	return nil
}
func (s *schemaTestRows) Next(values []driver.Value) error {
	if s.read {
		if s.mode == "row-error" {
			return errors.New("connection interrupted")
		}
		return io.EOF
	}
	s.read = true
	if s.mode == "scan-error" {
		values[0] = "not-an-int"
	} else {
		values[0] = int64(1)
	}
	return nil
}

var registerSchemaTestDriver sync.Once

func TestReadSchemaRowsRejectsIncompleteReads(t *testing.T) {
	registerSchemaTestDriver.Do(func() { sql.Register("magicorm-schema-test", schemaTestDriver{}) })
	for _, mode := range []string{"ok", "query-error", "row-error", "close-error", "scan-error", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			db, err := sql.Open("magicorm-schema-test", mode)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			result := ReadSchemaRows(ctx, db, "SELECT value", func(rows *sql.Rows) error {
				var value int
				err := rows.Scan(&value)
				if mode == "cancel" {
					cancel()
				}
				return err
			})
			if (result == nil) != (mode == "ok") {
				t.Fatalf("%s: %v", mode, result)
			}
			if mode == "row-error" && !strings.Contains(result.Error(), "incomplete") {
				t.Fatal(result)
			}
		})
	}
}
