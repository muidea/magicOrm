package mysql

import (
	"context"
	"database/sql"
	"regexp"
	"strings"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/models"
)

func (s *Builder) DescribeTable(model models.Model) (database.TableSchema, *cd.Error) {
	return database.DescribeModelTable(model, s.buildCodec, s.describeColumn)
}

func (s *Builder) describeColumn(field models.Field) (database.ColumnSchema, *cd.Error) {
	typeName, err := getTypeDeclare(field.GetType(), field.GetSpec())
	if err != nil {
		return database.ColumnSchema{}, err
	}
	defaultValue, err := s.validDefaultValue(field.GetType(), field.GetSpec())
	if err != nil {
		return database.ColumnSchema{}, err
	}
	auto, err := s.validAutoIncrement(field.GetType(), field.GetSpec())
	if err != nil {
		return database.ColumnSchema{}, err
	}
	if auto || defaultValue == "''" {
		defaultValue = ""
	}
	kind := normalizeSchemaType(typeName)
	return database.ColumnSchema{Name: field.GetName(), Type: kind, Nullable: field.GetType().IsPtrType() && !field.GetSpec().IsPrimaryKey(), AutoIncrement: auto, Default: database.NormalizeSchemaLiteral(defaultValue, kind)}, nil
}

var integerDisplayWidth = regexp.MustCompile(`^(tinyint|smallint|int|bigint)\([0-9]+\)$`)

func normalizeSchemaType(value string) string {
	value = strings.ToLower(value)
	// Integer display width is not storage size. Unsigned/zerofill declarations
	// deliberately do not match this pattern and remain drift.
	return integerDisplayWidth.ReplaceAllString(value, "$1")
}

func (s *Builder) DescribeRelationTable(model models.Model, field models.Field) (database.TableSchema, *cd.Error) {
	name, err := s.buildCodec.ConstructRelationTableName(model, field)
	if err != nil {
		return database.TableSchema{}, err
	}
	right, err := s.modelProvider.GetTypeModel(field.GetType().Elem())
	if err != nil {
		return database.TableSchema{}, err
	}
	leftType, err := getTypeDeclare(model.GetPrimaryField().GetType(), model.GetPrimaryField().GetSpec())
	if err != nil {
		return database.TableSchema{}, err
	}
	rightType, err := getTypeDeclare(right.GetPrimaryField().GetType(), right.GetPrimaryField().GetSpec())
	if err != nil {
		return database.TableSchema{}, err
	}
	return database.TableSchema{Name: name, Exists: true, Ordinary: true,
		Columns: []database.ColumnSchema{{Name: "id", Type: "bigint", AutoIncrement: true}, {Name: "left", Type: normalizeSchemaType(leftType)}, {Name: "right", Type: normalizeSchemaType(rightType)}},
		Indexes: []database.IndexSchema{{Name: "PRIMARY", Columns: []string{"id"}, Primary: true, Unique: true, Constraint: true, Valid: true}, {Name: "left", Columns: []string{"left"}, Valid: true}},
	}, nil
}

const schemaTableSQL = `SELECT TABLE_TYPE = 'BASE TABLE' FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?`
const schemaColumnsSQL = `SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE = 'YES', COLUMN_DEFAULT,
EXTRA LIKE '%auto_increment%', (EXTRA LIKE '%GENERATED%' OR EXTRA LIKE '%on update%' OR EXTRA LIKE '%INVISIBLE%')
FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? ORDER BY ORDINAL_POSITION`

// See https://dev.mysql.com/doc/refman/8.0/en/information-schema-statistics-table.html.
// A prefix, descending, invisible or functional index is not a full plain key.
const schemaIndexesSQL = `SELECT INDEX_NAME, COALESCE(COLUMN_NAME, ''), SEQ_IN_INDEX, NON_UNIQUE = 0, INDEX_NAME = 'PRIMARY', NON_UNIQUE = 0,
(INDEX_TYPE = 'BTREE' AND SUB_PART IS NULL AND COLLATION = 'A' AND IS_VISIBLE = 'YES' AND EXPRESSION IS NULL)
FROM information_schema.STATISTICS WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? ORDER BY INDEX_NAME, SEQ_IN_INDEX`

func readTableSchema(ctx context.Context, queryer database.SchemaQueryer, name string) (database.TableSchema, *cd.Error) {
	if name == "" {
		return database.TableSchema{}, cd.NewError(cd.IllegalParam, "schema inspection requires a table")
	}
	var schema sql.NullString
	if err := database.ReadSchemaRows(ctx, queryer, "SELECT DATABASE()", func(rows *sql.Rows) error { return rows.Scan(&schema) }); err != nil {
		return database.TableSchema{}, err
	}
	if !schema.Valid || schema.String == "" {
		return database.TableSchema{}, cd.NewError(cd.Unexpected, "schema inspection requires a selected database")
	}
	table, err := database.ReadCatalogTable(ctx, queryer, name, schemaTableSQL, schemaColumnsSQL, schemaIndexesSQL, schema.String, name)
	if err != nil {
		return database.TableSchema{}, err
	}
	for index := range table.Columns {
		column := &table.Columns[index]
		column.Type = normalizeSchemaType(column.Type)
		column.Default = database.NormalizeSchemaLiteral(column.Default, column.Type)
	}
	return table, nil
}

func (s *ConnExecutor) ReadTableSchema(ctx context.Context, name string) (database.TableSchema, *cd.Error) {
	if s.dbTx != nil {
		return readTableSchema(ctx, s.dbTx, name)
	}
	if s.dbConnPtr == nil {
		return database.TableSchema{}, cd.NewError(cd.Unexpected, "schema inspection connection unavailable")
	}
	return readTableSchema(ctx, s.dbConnPtr, name)
}

func (s *HostExecutor) ReadTableSchema(ctx context.Context, name string) (database.TableSchema, *cd.Error) {
	if s.dbTx != nil {
		return readTableSchema(ctx, s.dbTx, name)
	}
	if s.dbHandle == nil {
		return database.TableSchema{}, cd.NewError(cd.Unexpected, "schema inspection connection unavailable")
	}
	return readTableSchema(ctx, s.dbHandle, name)
}
