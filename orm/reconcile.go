package orm

import (
	"context"
	"fmt"
	"reflect"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

// ReconcileRunner evolves an already-created schema from a persisted model
// declaration to its replacement. It deliberately supports only additive
// operations; a destructive change requires an explicit deployment migration.
type ReconcileRunner struct {
	baseRunner
	previous models.Model
}

func NewReconcileRunner(ctx context.Context, previous, current models.Model, executor database.Executor, modelProvider provider.Provider, modelCodec codec.Codec) *ReconcileRunner {
	return &ReconcileRunner{
		baseRunner: newBaseRunner(ctx, current, executor, modelProvider, modelCodec, false, 0),
		previous:   previous,
	}
}

// fieldByName returns declared fields by their stable schema identifier. A
// missing old name is never interpreted as a rename.
func fieldByName(fields models.Fields) map[string]models.Field {
	ret := make(map[string]models.Field, len(fields))
	for _, field := range fields {
		if field != nil {
			ret[field.GetName()] = field
		}
	}
	return ret
}

func sameSchemaField(previous, current models.Field) bool {
	if previous == nil || current == nil || !models.CompareType(previous.GetType(), current.GetType()) {
		return false
	}
	if previous.GetSpec().IsPrimaryKey() != current.GetSpec().IsPrimaryKey() {
		return false
	}
	return true
}

// samePhysicalDefault compares backend declarations, not runtime default
// expressions. A declaration-only change needs no ALTER DEFAULT; a changed
// stored column contract still requires an explicit migration.
func (s *ReconcileRunner) samePhysicalDefault(previous, current models.Field) (bool, *cd.Error) {
	if !models.IsBasicField(previous) || !models.IsBasicField(current) {
		return false, nil
	}
	builder, ok := NewBuilder(s.modelProvider, s.modelCodec).(database.SchemaBuilder)
	if !ok {
		return false, cd.NewError(cd.Unexpected, "builder does not support schema declarations")
	}
	before, err := builder.DescribeTable(s.previous)
	if err != nil {
		return false, err
	}
	after, err := builder.DescribeTable(s.vModel)
	if err != nil {
		return false, err
	}
	for _, oldColumn := range before.Columns {
		if oldColumn.Name != previous.GetName() {
			continue
		}
		for _, newColumn := range after.Columns {
			if newColumn.Name == current.GetName() {
				return reflect.DeepEqual(oldColumn, newColumn), nil
			}
		}
	}
	return false, nil
}

func migrationRequired(format string, args ...any) *cd.Error {
	return cd.NewError(cd.InvalidOperation, "schema migration required: "+fmt.Sprintf(format, args...))
}

func findUniqueConstraint(items []models.UniqueConstraint, name string) (models.UniqueConstraint, bool) {
	for _, item := range items {
		if item.Name == name {
			return item, true
		}
	}
	return models.UniqueConstraint{}, false
}

func findIndex(items []models.Index, name string) (models.Index, bool) {
	for _, item := range items {
		if item.Name == name {
			return item, true
		}
	}
	return models.Index{}, false
}

func (s *ReconcileRunner) checkCompatibility() (map[string]models.Field, *cd.Error) {
	if s.previous == nil || s.vModel == nil {
		return nil, cd.NewError(cd.IllegalParam, "schema reconcile model is nil")
	}
	if s.previous.GetPkgKey() != s.vModel.GetPkgKey() {
		return nil, migrationRequired("entity identity changed from %q to %q", s.previous.GetPkgKey(), s.vModel.GetPkgKey())
	}

	previousFields := fieldByName(s.previous.GetFields())
	currentFields := fieldByName(s.vModel.GetFields())
	for name, previous := range previousFields {
		current, ok := currentFields[name]
		if !ok {
			return nil, migrationRequired("field %q was removed or renamed", name)
		}
		if !sameSchemaField(previous, current) {
			return nil, migrationRequired("field %q changed type, nullability, primary key, or default", name)
		}
		if !reflect.DeepEqual(previous.GetSpec().GetDefaultValue(), current.GetSpec().GetDefaultValue()) {
			same, err := s.samePhysicalDefault(previous, current)
			if err != nil {
				return nil, err
			}
			if !same {
				return nil, migrationRequired("field %q changed stored default", name)
			}
		}
	}

	previousPK := s.previous.GetPrimaryField()
	currentPK := s.vModel.GetPrimaryField()
	if previousPK == nil || currentPK == nil || previousPK.GetName() != currentPK.GetName() {
		return nil, migrationRequired("primary key changed")
	}

	for _, oldConstraint := range models.GetUniqueConstraints(s.previous) {
		current, ok := findUniqueConstraint(models.GetUniqueConstraints(s.vModel), oldConstraint.Name)
		if !ok || !reflect.DeepEqual(oldConstraint.Fields, current.Fields) {
			return nil, migrationRequired("unique constraint %q was removed or changed", oldConstraint.Name)
		}
	}
	for _, oldIndex := range models.GetIndexes(s.previous) {
		current, ok := findIndex(models.GetIndexes(s.vModel), oldIndex.Name)
		if !ok || !reflect.DeepEqual(oldIndex.Fields, current.Fields) {
			return nil, migrationRequired("index %q was removed or changed", oldIndex.Name)
		}
	}
	return previousFields, nil
}

func (s *ReconcileRunner) Reconcile() *cd.Error {
	statements, err := s.planReconcile()
	if err != nil {
		return err
	}
	return s.executeSchemaDDL(statements)
}

func (s *ReconcileRunner) planReconcile() ([]database.Result, *cd.Error) {
	if err := s.checkContext(); err != nil {
		return nil, err
	}
	previousFields, err := s.checkCompatibility()
	if err != nil {
		return nil, err
	}
	if _, err = collectSchemaGraph(s.context, s.previous, s.modelProvider, s.modelCodec); err != nil {
		return nil, err
	}
	if _, err = collectSchemaGraph(s.context, s.vModel, s.modelProvider, s.modelCodec); err != nil {
		return nil, err
	}
	statements := []database.Result{}
	created := map[string]bool{}

	for _, field := range s.vModel.GetFields() {
		if _, exists := previousFields[field.GetName()]; exists {
			continue
		}
		if models.IsBasicField(field) {
			// Existing rows cannot satisfy a new required column without a data
			// backfill. Nullable additions are the only universally lossless form.
			if !field.GetType().IsPtrType() {
				return nil, migrationRequired("new required field %q needs an explicit backfill", field.GetName())
			}
			result, buildErr := s.sqlBuilder.BuildAddColumn(s.vModel, field)
			if buildErr != nil {
				return nil, buildErr
			}
			statements = append(statements, result)
			continue
		}
		if !field.GetType().Elem().IsPtrType() {
			owned, err := s.modelProvider.GetTypeModel(field.GetType().Elem())
			if err != nil {
				return nil, err
			}
			graph, err := collectSchemaGraph(s.context, owned, s.modelProvider, s.modelCodec)
			if err != nil {
				return nil, err
			}
			// Multiple new fields may share the same owned model. Create each
			// host/relation once, while retaining both parent relation tables.
			filtered := &schemaGraph{}
			for _, step := range graph.order {
				key := step.model.GetPkgKey()
				if step.field != nil {
					key += "#" + step.field.GetName()
				}
				if !created[key] {
					filtered.order = append(filtered.order, step)
					created[key] = true
				}
			}
			ownedDDL, err := s.buildSchemaDDL(filtered, true)
			if err != nil {
				return nil, err
			}
			statements = append(statements, ownedDDL...)
		}

		result, buildErr := s.sqlBuilder.BuildCreateRelationTable(s.vModel, field)
		if buildErr != nil {
			return nil, buildErr
		}
		statements = append(statements, result)
	}

	previousConstraints := models.GetUniqueConstraints(s.previous)
	for _, constraint := range models.GetUniqueConstraints(s.vModel) {
		if _, exists := findUniqueConstraint(previousConstraints, constraint.Name); exists {
			continue
		}
		result, buildErr := s.sqlBuilder.BuildCreateUniqueConstraint(s.vModel, constraint)
		if buildErr != nil {
			return nil, buildErr
		}
		statements = append(statements, result)
	}

	previousIndexes := models.GetIndexes(s.previous)
	for _, index := range models.GetIndexes(s.vModel) {
		if _, exists := findIndex(previousIndexes, index.Name); exists {
			continue
		}
		result, buildErr := s.sqlBuilder.BuildCreateIndex(s.vModel, index)
		if buildErr != nil {
			return nil, buildErr
		}
		statements = append(statements, result)
	}
	return statements, nil
}

func (s *impl) Reconcile(previous, current models.Model) (err *cd.Error) {
	start := time.Now()
	defer func() {
		if ormMetricCollector != nil && current != nil {
			ormMetricCollector.RecordOperation("reconcile", current, time.Since(start), cd.ToStdError(err))
		}
	}()
	if err = s.CheckContext(); err != nil {
		return
	}
	if previous == nil || current == nil {
		return cd.NewError(cd.IllegalParam, "schema reconcile model is nil")
	}
	return NewReconcileRunner(s.context, previous, current, s.executor, s.modelProvider, s.modelCodec).Reconcile()
}
