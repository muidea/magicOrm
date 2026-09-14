package database

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"slices"
	"sort"
	"strconv"
	"strings"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
)

// Schema descriptors contain physical structure only, never connection credentials
// or row data. Default is a backend-normalized expression, not evaluated SQL.
type ColumnSchema struct {
	Name, Type, Default                string
	Nullable, AutoIncrement, Generated bool
}

type IndexSchema struct {
	Name                               string
	Columns                            []string
	Unique, Primary, Constraint, Valid bool
}

type TableSchema struct {
	Name             string
	Exists, Ordinary bool
	Columns          []ColumnSchema
	Indexes          []IndexSchema
}

// SchemaReader is a read-only capability, separate from the write executor.
type SchemaReader interface {
	ReadTableSchema(context.Context, string) (TableSchema, *cd.Error)
}

type SchemaBuilder interface {
	DescribeTable(models.Model) (TableSchema, *cd.Error)
	DescribeRelationTable(models.Model, models.Field) (TableSchema, *cd.Error)
}

type SchemaIssue struct {
	Table  string `json:"table"`
	Object string `json:"object,omitempty"`
	Kind   string `json:"kind"`
}

// NormalizeSchemaLiteral recognizes only scalar literals used by ORM DDL. SQL
// expressions are preserved verbatim and cannot be mistaken for a constant.
func NormalizeSchemaLiteral(value, kind string) string {
	original := value
	if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") && len(value) >= 2 {
		value = value[1 : len(value)-1]
	}
	switch kind {
	case "boolean":
		if value == "true" || value == "false" || value == "0" || value == "1" {
			return value
		}
	case "tinyint", "smallint", "integer", "int", "bigint":
		if integer, ok := new(big.Int).SetString(value, 10); ok {
			return integer.String()
		}
	case "real", "float", "double", "double precision":
		bits := 64
		if kind == "real" || kind == "float" {
			bits = 32
		}
		if number, err := strconv.ParseFloat(value, bits); err == nil {
			return strconv.FormatFloat(number, 'g', -1, bits)
		}
	}
	return original
}

// CompareTableSchema deliberately rejects extra columns and indexes too. It
// never proposes destructive repair or treats a same-name object as equivalent.
func CompareTableSchema(expected, actual TableSchema, exists bool) []SchemaIssue {
	issues := []SchemaIssue{}
	issue := func(object, kind string) { issues = append(issues, SchemaIssue{expected.Name, object, kind}) }
	if !exists {
		if actual.Exists {
			issue("", "table-still-exists")
		}
		return issues
	}
	if !actual.Exists {
		issue("", "table-missing")
		return issues
	}
	if !actual.Ordinary {
		issue("", "table-kind-mismatch")
		return issues
	}
	columns := map[string]ColumnSchema{}
	for _, column := range actual.Columns {
		if _, ok := columns[column.Name]; ok {
			issue(column.Name, "column-duplicate")
		}
		columns[column.Name] = column
	}
	for _, column := range expected.Columns {
		found, ok := columns[column.Name]
		if !ok {
			issue(column.Name, "column-missing")
		} else if column != found {
			issue(column.Name, "column-mismatch")
		}
		delete(columns, column.Name)
	}
	for name := range columns {
		issue(name, "column-unexpected")
	}
	indexes := map[string]IndexSchema{}
	for _, index := range actual.Indexes {
		if _, ok := indexes[index.Name]; ok {
			issue(index.Name, "index-duplicate")
		}
		indexes[index.Name] = index
	}
	for _, index := range expected.Indexes {
		found, ok := indexes[index.Name]
		if !ok {
			issue(index.Name, "index-missing")
		} else if !slices.Equal(index.Columns, found.Columns) || index.Unique != found.Unique || index.Primary != found.Primary || index.Constraint != found.Constraint || !found.Valid {
			issue(index.Name, "index-mismatch")
		}
		delete(indexes, index.Name)
	}
	for name := range indexes {
		issue(name, "index-unexpected")
	}
	sort.Slice(issues, func(i, j int) bool {
		if issues[i].Object == issues[j].Object {
			return issues[i].Kind < issues[j].Kind
		}
		return issues[i].Object < issues[j].Object
	})
	return issues
}

// DescribeModelTable shares model/index traversal; each SQL builder supplies
// column declarations using the same type/default helpers as its DDL builder.
func DescribeModelTable(model models.Model, modelCodec codec.Codec, column func(models.Field) (ColumnSchema, *cd.Error)) (TableSchema, *cd.Error) {
	if model == nil || model.GetPrimaryField() == nil {
		return TableSchema{}, cd.NewError(cd.IllegalParam, "schema model requires a primary key")
	}
	table := TableSchema{Name: modelCodec.ConstructModelTableName(model), Exists: true, Ordinary: true}
	for _, field := range model.GetFields() {
		if !models.IsBasicField(field) {
			continue
		}
		value, err := column(field)
		if err != nil {
			return TableSchema{}, err
		}
		table.Columns = append(table.Columns, value)
	}
	table.Indexes = append(table.Indexes, IndexSchema{Name: "PRIMARY", Columns: []string{model.GetPrimaryField().GetName()}, Unique: true, Primary: true, Constraint: true, Valid: true})
	for _, constraint := range models.GetUniqueConstraints(model) {
		if err := constraint.Verify(model.GetFields()); err != nil {
			return TableSchema{}, err
		}
		table.Indexes = append(table.Indexes, IndexSchema{Name: constraint.Name, Columns: slices.Clone(constraint.Fields), Unique: true, Constraint: true, Valid: true})
	}
	for _, index := range models.GetIndexes(model) {
		if err := index.Verify(model.GetFields()); err != nil {
			return TableSchema{}, err
		}
		table.Indexes = append(table.Indexes, IndexSchema{Name: index.Name, Columns: slices.Clone(index.Fields), Valid: true})
	}
	return table, nil
}

// SchemaQueryer is implemented by sql.DB, sql.Conn and sql.Tx. Reading catalog
// rows directly preserves Rows.Err; Executor.Next's boolean is not a receipt.
type SchemaQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func ReadSchemaRows(ctx context.Context, queryer SchemaQueryer, query string, scan func(*sql.Rows) error, args ...any) *cd.Error {
	if ctx == nil || ctx.Err() != nil {
		return cd.NewError(cd.Unexpected, "schema inspection context is invalid or cancelled")
	}
	rows, err := queryer.QueryContext(ctx, query, args...)
	if err != nil {
		return cd.NewError(cd.Unexpected, fmt.Sprintf("schema catalog query failed: %v", err))
	}
	defer rows.Close()
	for rows.Next() {
		if err := scan(rows); err != nil {
			return cd.NewError(cd.Unexpected, fmt.Sprintf("schema catalog row invalid: %v", err))
		}
	}
	if err := rows.Err(); err != nil {
		return cd.NewError(cd.Unexpected, fmt.Sprintf("schema catalog read incomplete: %v", err))
	}
	if err := rows.Close(); err != nil {
		return cd.NewError(cd.Unexpected, fmt.Sprintf("schema catalog close failed: %v", err))
	}
	if ctx.Err() != nil {
		return cd.NewError(cd.Unexpected, "schema inspection context cancelled")
	}
	return nil
}

// ReadCatalogTable uses backend queries with a common projection. All parameters
// are bound values, including table and schema names. Errors discard partial data.
func ReadCatalogTable(ctx context.Context, queryer SchemaQueryer, tableName, tableSQL, columnSQL, indexSQL string, args ...any) (TableSchema, *cd.Error) {
	table := TableSchema{Name: tableName}
	err := ReadSchemaRows(ctx, queryer, tableSQL, func(rows *sql.Rows) error {
		if table.Exists {
			return fmt.Errorf("duplicate table")
		}
		table.Exists = true
		return rows.Scan(&table.Ordinary)
	}, args...)
	if err != nil {
		return TableSchema{}, err
	}
	if !table.Exists || !table.Ordinary {
		return table, nil
	}
	err = ReadSchemaRows(ctx, queryer, columnSQL, func(rows *sql.Rows) error {
		column := ColumnSchema{}
		var defaultValue sql.NullString
		if err := rows.Scan(&column.Name, &column.Type, &column.Nullable, &defaultValue, &column.AutoIncrement, &column.Generated); err != nil {
			return err
		}
		column.Default = defaultValue.String
		table.Columns = append(table.Columns, column)
		return nil
	}, args...)
	if err != nil {
		return TableSchema{}, err
	}
	if len(table.Columns) == 0 {
		return TableSchema{}, cd.NewError(cd.Unexpected, "schema catalog columns unavailable")
	}
	indexPositions := map[string]int{}
	err = ReadSchemaRows(ctx, queryer, indexSQL, func(rows *sql.Rows) error {
		index := IndexSchema{}
		var column string
		var position int
		if err := rows.Scan(&index.Name, &column, &position, &index.Unique, &index.Primary, &index.Constraint, &index.Valid); err != nil {
			return err
		}
		if index.Primary {
			index.Name = "PRIMARY"
		}
		if pos, ok := indexPositions[index.Name]; ok {
			previous := &table.Indexes[pos]
			if position != len(previous.Columns)+1 || previous.Unique != index.Unique || previous.Primary != index.Primary || previous.Constraint != index.Constraint {
				return fmt.Errorf("incomplete or duplicate index columns")
			}
			previous.Columns = append(previous.Columns, column)
			previous.Valid = previous.Valid && index.Valid
		} else {
			if position != 1 {
				return fmt.Errorf("index first column missing")
			}
			index.Columns = []string{column}
			indexPositions[index.Name] = len(table.Indexes)
			table.Indexes = append(table.Indexes, index)
		}
		return nil
	}, args...)
	if err != nil {
		return TableSchema{}, err
	}
	return table, nil
}
