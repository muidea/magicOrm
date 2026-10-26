package orm

import (
	"context"
	"fmt"
	"sort"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/internal/database"
	"github.com/muidea/magicOrm/internal/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

const maxSchemaTables = 1024

// SchemaTableName observes the host table identity only, without opening a
// connection or granting write authority. Unlike SchemaTables it excludes
// relation tables and recursively owned models.
func SchemaTableName(model models.Model, prefix string) (string, *cd.Error) {
	if model == nil || model.GetName() == "" {
		return "", cd.NewError(cd.IllegalParam, "schema model identity is unavailable")
	}
	return codec.New(nil, prefix).ConstructModelTableName(model), nil
}

// schemaTable retains the declaration used to build each statement. The graph
// is validated in full before any DDL is built or executed.
type schemaTable struct {
	model models.Model
	field models.Field // nil for a host table
}

type schemaGraph struct {
	tables map[string]database.TableSchema
	order  []schemaTable
}

// ValidateSchema validates a declaration without opening a database connection.
// It does not establish that the physical schema exists or matches the model.
func ValidateSchema(ctx context.Context, model models.Model, p provider.Provider, prefix string) *cd.Error {
	graph, err := collectSchemaGraph(ctx, model, p, codec.New(p, prefix))
	if err != nil {
		return err
	}
	return validateSchemaIndexes(graph)
}

func validateSchemaIndexes(graph *schemaGraph) *cd.Error {
	for _, step := range graph.order {
		if step.field == nil {
			if err := models.ValidateIndexes(step.model); err != nil {
				return err
			}
		}
	}
	return nil
}

// InspectSchema checks exactly the tables Create/Drop own: the host, relation
// tables, and recursively owned models, not the hosts of referenced models.
// The caller must serialize schema writers; this is a read-only observation,
// not a distributed DDL lock or a repair command. A read error returns no report.
func (s *impl) InspectSchema(model models.Model, expectedExist bool) (*SchemaReport, *cd.Error) {
	if err := s.checkContext(); err != nil {
		return nil, err
	}
	reader, ok := s.executor.(database.SchemaReader)
	if !ok {
		return nil, cd.NewError(cd.Unexpected, "executor does not support physical schema inspection")
	}
	graph, err := collectSchemaGraph(s.context, model, s.modelProvider, s.modelCodec)
	if err != nil {
		return nil, err
	}
	report := &SchemaReport{Tables: make([]string, 0, len(graph.tables)), Issues: []database.SchemaIssue{}}
	for name := range graph.tables {
		report.Tables = append(report.Tables, name)
	}
	sort.Strings(report.Tables)
	for _, name := range report.Tables {
		if err := s.checkContext(); err != nil {
			return nil, err
		}
		actual, err := reader.ReadTableSchema(s.context, name)
		if err != nil {
			return nil, err
		}
		if actual.Name != name {
			return nil, cd.NewError(cd.Unexpected, fmt.Sprintf("schema inspection receipt table mismatch: %s", name))
		}
		report.Issues = append(report.Issues, database.CompareTableSchema(graph.tables[name], actual, expectedExist)...)
	}
	if err := s.checkContext(); err != nil {
		return nil, err
	}
	report.Matches = len(report.Issues) == 0
	return report, nil
}

func collectSchemaGraph(ctx context.Context, model models.Model, p provider.Provider, c codec.Codec) (*schemaGraph, *cd.Error) {
	if !isContextValid(ctx) {
		return nil, cd.NewError(cd.Unexpected, "context is invalid or cancelled")
	}
	if model == nil || p == nil || c == nil {
		return nil, cd.NewError(cd.IllegalParam, "schema model/provider unavailable")
	}
	builder, ok := newBuilder(p, c).(database.SchemaBuilder)
	if !ok {
		return nil, cd.NewError(cd.Unexpected, "builder does not support schema declarations")
	}
	graph := &schemaGraph{tables: map[string]database.TableSchema{}}
	visited := map[string]bool{}
	active := map[string]bool{}
	add := func(table database.TableSchema, step schemaTable) *cd.Error {
		if table.Name == "" || len(graph.tables) >= maxSchemaTables {
			return cd.NewError(cd.IllegalParam, "schema table limit or invalid name")
		}
		if _, exists := graph.tables[table.Name]; exists {
			return cd.NewError(cd.IllegalParam, "schema physical table name collision")
		}
		graph.tables[table.Name] = table
		graph.order = append(graph.order, step)
		return nil
	}
	var visit func(models.Model) *cd.Error
	visit = func(value models.Model) *cd.Error {
		if !isContextValid(ctx) {
			return cd.NewError(cd.Unexpected, "context is invalid or cancelled")
		}
		if value == nil || value.GetPrimaryField() == nil {
			return cd.NewError(cd.IllegalParam, "schema model requires a primary key")
		}
		key := value.GetPkgKey()
		if active[key] {
			return cd.NewError(cd.IllegalParam, "schema owned model cycle")
		}
		if visited[key] {
			return nil
		}
		active[key] = true
		table, err := builder.DescribeTable(value)
		if err != nil {
			return err
		}
		if err := add(table, schemaTable{model: value}); err != nil {
			return err
		}
		for _, field := range value.GetFields() {
			if models.IsMainTableField(value, field) {
				continue
			}
			right, err := p.GetTypeModel(field.GetType().Elem())
			if err != nil {
				return err
			}
			if right == nil || right.GetPrimaryField() == nil {
				return cd.NewError(cd.IllegalParam, "schema inspection relation model unavailable")
			}
			if !field.GetType().Elem().IsPtrType() {
				if err := visit(right); err != nil {
					return err
				}
			}
			relation, err := builder.DescribeRelationTable(value, field)
			if err != nil {
				return err
			}
			if err := add(relation, schemaTable{model: value, field: field}); err != nil {
				return err
			}
		}
		active[key], visited[key] = false, true
		return nil
	}
	if err := visit(model); err != nil {
		return nil, err
	}
	if !isContextValid(ctx) {
		return nil, cd.NewError(cd.Unexpected, "context is invalid or cancelled")
	}
	return graph, nil
}

// SchemaTables returns the validated physical table identities managed by
// Create/Drop, without accessing a database. It conveys no ownership authority.
func SchemaTables(ctx context.Context, model models.Model, p provider.Provider, prefix string) ([]string, *cd.Error) {
	graph, err := collectSchemaGraph(ctx, model, p, codec.New(p, prefix))
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(graph.tables))
	for name := range graph.tables {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}
