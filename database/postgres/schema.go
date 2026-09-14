package postgres

import (
	"context"
	"strings"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/models"
)

func (s *Builder) DescribeTable(model models.Model) (database.TableSchema, *cd.Error) {
	return database.DescribeModelTable(model, s.buildCodec, s.describeColumn)
}

func (s *Builder) describeColumn(field models.Field) (database.ColumnSchema, *cd.Error) {
	typeName, err := getTypeDeclare(field.GetType(), field.GetSpec(), true)
	if err != nil {
		return database.ColumnSchema{}, err
	}
	defaultValue, err := s.validDefaultValue(field.GetType(), field.GetSpec())
	if err != nil {
		return database.ColumnSchema{}, err
	}
	auto := typeName == "SERIAL" || typeName == "BIGSERIAL" || typeName == "SMALLSERIAL"
	autoDeclare, err := s.validAutoIncrement(field.GetType(), field.GetSpec())
	if err != nil {
		return database.ColumnSchema{}, err
	}
	if autoDeclare || defaultValue == "''" {
		defaultValue = ""
	}
	kind := normalizeSchemaType(typeName)
	return database.ColumnSchema{Name: field.GetName(), Type: kind, Nullable: field.GetType().IsPtrType() && !field.GetSpec().IsPrimaryKey(), AutoIncrement: auto, Default: normalizeSchemaDefault(defaultValue, kind)}, nil
}

func normalizeSchemaType(value string) string {
	value = strings.ToLower(value)
	switch value {
	case "serial":
		return "integer"
	case "bigserial":
		return "bigint"
	case "smallserial":
		return "smallint"
	case "varchar(32)":
		return "character varying(32)"
	case "timestamp(3)":
		return "timestamp(3) without time zone"
	}
	return value
}

func normalizeSchemaDefault(value, kind string) string {
	// Only remove casts emitted for the expected physical type. Unknown
	// expressions remain unequal; never execute catalog expressions as SQL.
	cast := kind
	if strings.HasPrefix(kind, "timestamp(") {
		cast = "timestamp without time zone"
	}
	value = strings.TrimSuffix(value, "::"+cast)
	value = database.NormalizeSchemaLiteral(value, kind)
	if kind == "boolean" {
		if value == "1" {
			return "true"
		}
		if value == "0" {
			return "false"
		}
	}
	return value
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
	leftType, err := getTypeDeclare(model.GetPrimaryField().GetType(), model.GetPrimaryField().GetSpec(), false)
	if err != nil {
		return database.TableSchema{}, err
	}
	rightType, err := getTypeDeclare(right.GetPrimaryField().GetType(), right.GetPrimaryField().GetSpec(), false)
	if err != nil {
		return database.TableSchema{}, err
	}
	return database.TableSchema{Name: name, Exists: true, Ordinary: true,
		Columns: []database.ColumnSchema{{Name: "id", Type: "bigint", AutoIncrement: true}, {Name: "left", Type: normalizeSchemaType(leftType)}, {Name: "right", Type: normalizeSchemaType(rightType)}},
		Indexes: []database.IndexSchema{{Name: "PRIMARY", Columns: []string{"id"}, Primary: true, Unique: true, Constraint: true, Valid: true}, {Name: name + "_index", Columns: []string{"left"}, Valid: true}},
	}, nil
}

// All catalog queries bind schema and table separately. No search_path fallback
// or similarly named object in another schema can satisfy the receipt.
const schemaTableSQL = `SELECT c.relkind = 'r' FROM pg_catalog.pg_class c
JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname = $1 AND c.relname = $2`

const schemaColumnsSQL = `SELECT a.attname, pg_catalog.format_type(a.atttypid, a.atttypmod), NOT a.attnotnull,
pg_catalog.pg_get_expr(d.adbin, d.adrelid),
(a.attidentity <> '' OR (pg_catalog.pg_get_expr(d.adbin, d.adrelid) LIKE 'nextval(%::regclass)' AND EXISTS (
 SELECT 1 FROM pg_catalog.pg_depend dep WHERE dep.classid = 'pg_catalog.pg_attrdef'::regclass AND dep.objid = d.oid
 AND dep.refclassid = 'pg_catalog.pg_class'::regclass AND dep.refobjid = pg_catalog.pg_get_serial_sequence(format('%I.%I', n.nspname, c.relname), a.attname)::regclass))) IS TRUE,
(a.attgenerated <> '' OR a.attidentity <> '')
FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
JOIN pg_catalog.pg_attribute a ON a.attrelid = c.oid
LEFT JOIN pg_catalog.pg_attrdef d ON d.adrelid = c.oid AND d.adnum = a.attnum
WHERE n.nspname = $1 AND c.relname = $2 AND a.attnum > 0 AND NOT a.attisdropped ORDER BY a.attnum`

// See https://www.postgresql.org/docs/current/catalog-pg-index.html.
// Invalid/partial/expression/include/custom operator-class indexes do not count
// as the plain immediate btree indexes emitted by this builder.
const schemaIndexesSQL = `SELECT ic.relname, COALESCE(a.attname, ''), k.ordinality, i.indisunique, i.indisprimary,
EXISTS (SELECT 1 FROM pg_catalog.pg_constraint con WHERE con.conindid = i.indexrelid AND con.conrelid = c.oid AND con.contype IN ('p','u') AND NOT con.condeferrable),
(i.indisvalid AND i.indisready AND i.indislive AND i.indimmediate AND NOT i.indisexclusion
 AND i.indpred IS NULL AND i.indexprs IS NULL AND i.indnkeyatts = i.indnatts AND am.amname = 'btree'
 AND i.indoption[k.ordinality-1] = 0 AND opc.opcdefault AND i.indcollation[k.ordinality-1] = a.attcollation
 AND NOT COALESCE((to_jsonb(i)->>'indnullsnotdistinct')::boolean, false)) IS TRUE
FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
JOIN pg_catalog.pg_index i ON i.indrelid = c.oid JOIN pg_catalog.pg_class ic ON ic.oid = i.indexrelid
JOIN pg_catalog.pg_am am ON am.oid = ic.relam
CROSS JOIN LATERAL unnest(i.indkey) WITH ORDINALITY k(attnum, ordinality)
LEFT JOIN pg_catalog.pg_attribute a ON a.attrelid = c.oid AND a.attnum = k.attnum
LEFT JOIN pg_catalog.pg_opclass opc ON opc.oid = i.indclass[k.ordinality-1]
WHERE n.nspname = $1 AND c.relname = $2 ORDER BY ic.relname, k.ordinality`

func readTableSchema(ctx context.Context, queryer database.SchemaQueryer, schema, name string) (database.TableSchema, *cd.Error) {
	if schema == "" || name == "" {
		return database.TableSchema{}, cd.NewError(cd.IllegalParam, "schema inspection requires an explicit schema and table")
	}
	table, err := database.ReadCatalogTable(ctx, queryer, name, schemaTableSQL, schemaColumnsSQL, schemaIndexesSQL, schema, name)
	if err != nil {
		return database.TableSchema{}, err
	}
	for index := range table.Columns {
		column := &table.Columns[index]
		column.Type = normalizeSchemaType(column.Type)
		if column.AutoIncrement {
			column.Default = ""
		} else {
			column.Default = normalizeSchemaDefault(column.Default, column.Type)
		}
	}
	return table, nil
}

func (s *ConnExecutor) ReadTableSchema(ctx context.Context, name string) (database.TableSchema, *cd.Error) {
	if s.dbTx != nil {
		return readTableSchema(ctx, s.dbTx, s.schemaName, name)
	}
	if s.dbConnPtr == nil {
		return database.TableSchema{}, cd.NewError(cd.Unexpected, "schema inspection connection unavailable")
	}
	return readTableSchema(ctx, s.dbConnPtr, s.schemaName, name)
}

func (s *HostExecutor) ReadTableSchema(ctx context.Context, name string) (database.TableSchema, *cd.Error) {
	if s.dbTx != nil {
		return readTableSchema(ctx, s.dbTx, s.schemaName, name)
	}
	if s.dbHandle == nil {
		return database.TableSchema{}, cd.NewError(cd.Unexpected, "schema inspection connection unavailable")
	}
	return readTableSchema(ctx, s.dbHandle, s.schemaName, name)
}
