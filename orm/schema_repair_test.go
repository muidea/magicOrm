package orm

import (
	"context"
	"maps"
	"slices"
	"strings"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
)

type repairExecutor struct {
	*schemaExecutor
	apply func(string)
	fail  bool
}

type repairIndexedModel struct {
	models.Model
	indexes     []models.Index
	constraints []models.UniqueConstraint
}

func (m repairIndexedModel) GetIndexes() []models.Index                      { return m.indexes }
func (m repairIndexedModel) GetUniqueConstraints() []models.UniqueConstraint { return m.constraints }

func indexedRepairModel(model models.Model) models.Model {
	return repairIndexedModel{Model: model,
		indexes:     []models.Index{{Name: "repair_namespace_index", Fields: []string{"namespace"}}},
		constraints: []models.UniqueConstraint{{Name: "repair_audience_unique", Fields: []string{"audience"}}},
	}
}

func TestSchemaRepairReplaysOnlyMissingIndexAndUniqueSteps(t *testing.T) {
	for _, alreadyUnique := range []bool{false, true} {
		p, previous, desired := reconcileModels(t, &reconcileCurrentModel{})
		desired = indexedRepairModel(desired)
		c := codec.New(p, "tenant")
		target, err := NewBuilder(p, c).(database.SchemaBuilder).DescribeTable(desired)
		if err != nil {
			t.Fatal(err)
		}
		partial := target
		count := 2
		partial.Indexes = slices.Clone(target.Indexes[:1])
		if alreadyUnique {
			partial.Indexes = slices.Clone(target.Indexes[:2])
			count = 1
		}
		executor := &repairExecutor{schemaExecutor: &schemaExecutor{tables: map[string]database.TableSchema{target.Name: partial}}}
		executor.apply = func(sql string) {
			if strings.Contains(sql, "ADD COLUMN") || (alreadyUnique && strings.Contains(sql, "ADD CONSTRAINT")) {
				t.Fatal("replayed completed step", sql)
			}
			if len(executor.execCalls) == count {
				executor.tables[target.Name] = target
			}
		}
		handler := &impl{context: context.Background(), executor: executor, modelProvider: p, modelCodec: c}
		if report, err := handler.RepairSchema(previous, desired); err != nil || report == nil || !report.Matches || len(executor.execCalls) != count {
			t.Fatal(report, err, executor.execCalls)
		}
	}
}

func (s *repairExecutor) Execute(sql string, args ...any) (int64, *cd.Error) {
	n, err := s.fakeExecutor.Execute(sql, args...)
	if s.apply != nil {
		s.apply(sql)
	}
	if s.fail {
		return 0, cd.NewError(cd.DatabaseError, "ambiguous DDL")
	}
	return n, err
}

func TestSchemaRepairCompletesMissingTablesAndSkipsFinishedSteps(t *testing.T) {
	for _, drop := range []bool{false, true} {
		t.Run(map[bool]string{true: "drop", false: "create"}[drop], func(t *testing.T) {
			handler, model, catalog := inspectionFixture(t)
			expected := maps.Clone(catalog.tables)
			keys := slices.Sorted(maps.Keys(expected))
			delete(catalog.tables, keys[0])
			delete(catalog.tables, keys[1])
			executor := &repairExecutor{schemaExecutor: catalog}
			executor.apply = func(sql string) {
				for name, table := range expected {
					if strings.Contains(sql, dialectSQLForTest(`"`+name+`"`)) {
						if drop {
							delete(catalog.tables, name)
						} else {
							catalog.tables[name] = table
						}
					}
				}
			}
			handler.executor = executor
			var before, after models.Model = nil, model
			if drop {
				before, after = model, nil
			}
			report, err := handler.RepairSchema(before, after)
			if err != nil || report == nil || !report.Matches || len(executor.execCalls) != 2 {
				t.Fatal(report, err, executor.execCalls)
			}
			for _, call := range executor.execCalls {
				if strings.Contains(call.sql, "InspectReferenced") && !strings.Contains(call.sql, "InspectHost") {
					t.Fatal("repaired unowned reference host")
				}
			}
			report, err = handler.RepairSchema(before, after)
			if err != nil || !report.Matches || len(executor.execCalls) != 2 {
				t.Fatal("completed steps were replayed", report, err)
			}
		})
	}
}

func TestSchemaRepairReconcileResumesOnlyMissingNullableAddition(t *testing.T) {
	for _, ambiguous := range []bool{false, true} {
		p, before, after := reconcileModels(t, &reconcileCurrentModel{})
		c := codec.New(p, "tenant")
		builder := NewBuilder(p, c).(database.SchemaBuilder)
		old, err := builder.DescribeTable(before)
		if err != nil {
			t.Fatal(err)
		}
		target, err := builder.DescribeTable(after)
		if err != nil {
			t.Fatal(err)
		}
		executor := &repairExecutor{schemaExecutor: &schemaExecutor{tables: map[string]database.TableSchema{old.Name: old}}, fail: ambiguous}
		executor.apply = func(sql string) {
			if !strings.Contains(sql, "ADD COLUMN") || !strings.Contains(sql, "audience") {
				t.Fatal(sql)
			}
			executor.tables[old.Name] = target
		}
		handler := &impl{context: context.Background(), executor: executor, modelProvider: p, modelCodec: c}
		report, repairErr := handler.RepairSchema(before, after)
		if ambiguous {
			if repairErr == nil || report != nil {
				t.Fatal(report, repairErr)
			}
			executor.fail = false
			report, repairErr = handler.RepairSchema(before, after)
		}
		if repairErr != nil || report == nil || !report.Matches || len(executor.execCalls) != 1 {
			t.Fatal(report, repairErr, executor.execCalls)
		}
	}
}

func TestSchemaRepairRejectsDriftAnywhereBeforeAnyWrite(t *testing.T) {
	for _, mode := range []string{"column-mismatch", "column-missing", "column-extra", "primary-missing", "index-mismatch", "view", "catalog-error", "wrong-name", "canceled"} {
		t.Run(mode, func(t *testing.T) {
			handler, model, executor := inspectionFixture(t)
			graph, err := collectSchemaGraph(handler.context, model, handler.modelProvider, handler.modelCodec)
			if err != nil {
				t.Fatal(err)
			}
			host := handler.modelCodec.ConstructModelTableName(model)
			delete(executor.tables, host) // a valid early create must not execute
			last := graph.order[len(graph.order)-1]
			name, err := handler.modelCodec.ConstructRelationTableName(last.model, last.field)
			if err != nil {
				t.Fatal(err)
			}
			table := executor.tables[name]
			switch mode {
			case "column-mismatch":
				table.Columns[0].Type = "unexpected"
			case "column-missing":
				table.Columns = table.Columns[1:]
			case "column-extra":
				table.Columns = append(table.Columns, database.ColumnSchema{Name: "extra"})
			case "primary-missing":
				table.Indexes = table.Indexes[1:]
			case "index-mismatch":
				table.Indexes[1].Valid = false
			case "view":
				table.Ordinary = false
			case "catalog-error":
				executor.err = cd.NewError(cd.DatabaseError, "read failed")
			case "wrong-name":
				table.Name = "foreign"
			case "canceled":
				var cancel context.CancelFunc
				handler.context, cancel = context.WithCancel(handler.context)
				defer cancel()
				executor.cancel = cancel
			}
			executor.tables[name] = table
			if report, err := handler.RepairSchema(nil, model); err == nil || report != nil || len(executor.execCalls) != 0 {
				t.Fatal("drift produced partial DDL", report, err, executor.execCalls)
			}
		})
	}
}

func TestSchemaRepairRefusesLossOfPreexistingDataAndRejectsDestructivePlan(t *testing.T) {
	for _, mode := range []string{"missing-table", "missing-old-column", "missing-old-primary", "invalid-addition", "destructive"} {
		t.Run(mode, func(t *testing.T) {
			var target any = &reconcileCurrentModel{}
			if mode == "invalid-addition" {
				target = &reconcileLateRequired{}
			}
			if mode == "destructive" {
				target = &reconcileDestructiveModel{}
			}
			p, before, after := reconcileModels(t, target)
			c := codec.New(p, "tenant")
			table, err := NewBuilder(p, c).(database.SchemaBuilder).DescribeTable(before)
			if err != nil {
				t.Fatal(err)
			}
			switch mode {
			case "missing-table":
				table.Exists = false
			case "missing-old-column":
				table.Columns = table.Columns[:1]
			case "missing-old-primary":
				table.Indexes = nil
			}
			executor := &schemaExecutor{tables: map[string]database.TableSchema{table.Name: table}}
			handler := &impl{context: context.Background(), executor: executor, modelProvider: p, modelCodec: c}
			if report, err := handler.RepairSchema(before, after); err == nil || report != nil || len(executor.execCalls) != 0 {
				t.Fatal(report, err)
			}
		})
	}
}

func TestSchemaRepairMustConfirmPhysicalResult(t *testing.T) {
	handler, model, executor := inspectionFixture(t)
	executor.tables = map[string]database.TableSchema{} // SQL echo does not update catalogs
	if report, err := handler.RepairSchema(nil, model); err == nil || report != nil || len(executor.execCalls) != 4 {
		t.Fatal(report, err)
	}
}
