package orm

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/models"
)

// RepairSchema completes only missing steps of a caller-owned, frozen create,
// additive reconcile or drop. The caller must authenticate the operation and
// serialize all schema writers. It is not automatic migration or a DDL lock.
// Existing tables/columns/indexes that differ from the declaration are rejected
// before any write; missing pre-existing data tables are never recreated.
func (s *impl) RepairSchema(previous, desired models.Model) (*SchemaReport, *cd.Error) {
	if err := s.CheckContext(); err != nil {
		return nil, err
	}
	reader, ok := s.executor.(database.SchemaReader)
	if !ok {
		return nil, cd.NewError(cd.Unexpected, "executor does not support physical schema inspection")
	}
	target := desired
	if target == nil {
		target = previous
	}
	runner := newBaseRunner(s.context, target, s.executor, s.modelProvider, s.modelCodec, false, 0)
	graph, err := collectSchemaGraph(s.context, target, s.modelProvider, s.modelCodec)
	if err != nil {
		return nil, err
	}
	var baseline *schemaGraph
	if previous != nil && desired != nil {
		if _, err := NewReconcileRunner(s.context, previous, desired, nil, s.modelProvider, s.modelCodec).planReconcile(); err != nil {
			return nil, err
		}
		baseline, err = collectSchemaGraph(s.context, previous, s.modelProvider, s.modelCodec)
	} else {
		_, err = runner.buildSchemaDDL(graph, desired != nil)
	}
	if err != nil {
		return nil, err
	}
	statements, err := runner.planSchemaRepair(graph, baseline, reader, desired == nil)
	if err != nil {
		return nil, err
	}
	if err := runner.executeSchemaDDL(statements); err != nil {
		return nil, err
	}
	report, err := s.InspectSchema(target, desired != nil)
	if err != nil {
		return nil, err
	}
	if !report.Matches {
		return nil, cd.NewError(cd.DataCorrupted, "schema repair physical confirmation failed")
	}
	return report, nil
}

func (s *baseRunner) planSchemaRepair(graph, baseline *schemaGraph, reader database.SchemaReader, drop bool) ([]database.Result, *cd.Error) {
	statements := []database.Result{}
	for _, step := range graph.order {
		if err := s.checkContext(); err != nil {
			return nil, err
		}
		name := s.modelCodec.ConstructModelTableName(step.model)
		if step.field != nil {
			var err *cd.Error
			name, err = s.modelCodec.ConstructRelationTableName(step.model, step.field)
			if err != nil {
				return nil, err
			}
		}
		expected := graph.tables[name]
		actual, err := reader.ReadTableSchema(s.context, name)
		if err != nil {
			return nil, err
		}
		if actual.Name != name {
			return nil, cd.NewError(cd.DataCorrupted, "schema repair catalog scope mismatch")
		}
		var before database.TableSchema
		if baseline != nil {
			before = baseline.tables[name]
		}
		issues := database.CompareTableSchema(expected, actual, true)
		if drop {
			if !actual.Exists {
				continue
			}
			if len(issues) != 0 {
				return nil, schemaRepairDrift(issues[0])
			}
		} else if actual.Exists {
			missing, err := s.repairMissingObjects(step, before, issues)
			if err != nil {
				return nil, err
			}
			statements = append(statements, missing...)
			continue
		} else if before.Exists {
			return nil, cd.NewError(cd.InvalidOperation, "schema repair refuses to recreate a pre-existing data table")
		}
		// A completely missing new table or an exact remaining drop target.
		ddl, err := s.buildSchemaDDL(&schemaGraph{order: []schemaTable{step}}, !drop)
		if err != nil {
			return nil, err
		}
		statements = append(statements, ddl...)
	}
	return statements, s.checkContext()
}

func schemaRepairDrift(issue database.SchemaIssue) *cd.Error {
	return cd.NewError(cd.InvalidOperation, fmt.Sprintf("schema repair refuses drift: table=%s object=%s kind=%s", issue.Table, issue.Object, issue.Kind))
}

// SQL comes only from the frozen model's builder. Catalog names/defaults are
// evidence for comparison, never executable input. Missing objects that were
// already present in the old declaration are drift, not interrupted additions.
func (s *baseRunner) repairMissingObjects(step schemaTable, before database.TableSchema, issues []database.SchemaIssue) ([]database.Result, *cd.Error) {
	columns, indexes := map[string]bool{}, map[string]bool{}
	for _, issue := range issues {
		switch issue.Kind {
		case "column-missing":
			if !before.Exists || step.field != nil {
				return nil, schemaRepairDrift(issue)
			}
			for _, column := range before.Columns {
				if column.Name == issue.Object {
					return nil, schemaRepairDrift(issue)
				}
			}
			columns[issue.Object] = true
		case "index-missing":
			if step.field != nil || issue.Object == "PRIMARY" {
				return nil, schemaRepairDrift(issue)
			}
			for _, index := range before.Indexes {
				if index.Name == issue.Object {
					return nil, schemaRepairDrift(issue)
				}
			}
			indexes[issue.Object] = true
		default:
			return nil, schemaRepairDrift(issue)
		}
	}
	statements := []database.Result{}
	for _, field := range step.model.GetFields() {
		if !columns[field.GetName()] {
			continue
		}
		if !models.IsBasicField(field) || !field.GetType().IsPtrType() || field.GetSpec().IsPrimaryKey() {
			return nil, cd.NewError(cd.InvalidOperation, "schema repair only adds new nullable fields")
		}
		statement, err := s.sqlBuilder.BuildAddColumn(step.model, field)
		if err != nil {
			return nil, err
		}
		statements = append(statements, statement)
		delete(columns, field.GetName())
	}
	for _, constraint := range models.GetUniqueConstraints(step.model) {
		if indexes[constraint.Name] {
			statement, err := s.sqlBuilder.BuildCreateUniqueConstraint(step.model, constraint)
			if err != nil {
				return nil, err
			}
			statements = append(statements, statement)
			delete(indexes, constraint.Name)
		}
	}
	for _, index := range models.GetIndexes(step.model) {
		if indexes[index.Name] {
			statement, err := s.sqlBuilder.BuildCreateIndex(step.model, index)
			if err != nil {
				return nil, err
			}
			statements = append(statements, statement)
			delete(indexes, index.Name)
		}
	}
	if len(columns)+len(indexes) != 0 {
		return nil, cd.NewError(cd.InvalidOperation, "schema repair cannot map missing objects to frozen declarations")
	}
	return statements, nil
}
