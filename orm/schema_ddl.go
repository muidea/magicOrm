package orm

import (
	"context"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

// PreflightSchemaChange checks the entire declaration and builds all DDL,
// without opening a connection or executing SQL. nil previous means create;
// nil desired means drop; both non-nil mean additive reconcile. The provider
// must contain the related declarations for the intended operation. This is
// not a physical completion receipt and callers must still inspect after DDL.
func PreflightSchemaChange(ctx context.Context, previous, desired models.Model, p provider.Provider, prefix string) *cd.Error {
	c := codec.New(p, prefix)
	if previous != nil && desired != nil {
		_, err := NewReconcileRunner(ctx, previous, desired, nil, p, c).planReconcile()
		return err
	}
	model := desired
	if model == nil {
		model = previous
	}
	graph, err := collectSchemaGraph(ctx, model, p, c)
	if err != nil {
		return err
	}
	runner := newBaseRunner(ctx, model, nil, p, c, false, 0)
	_, err = runner.buildSchemaDDL(graph, desired != nil)
	return err
}

// buildSchemaDDL materializes every statement before the executor sees any of
// them. A late builder/declaration failure must not leave an early table behind.
func (s *baseRunner) buildSchemaDDL(graph *schemaGraph, create bool) ([]database.Result, *cd.Error) {
	statements := make([]database.Result, 0, len(graph.order))
	for _, step := range graph.order {
		if err := s.checkContext(); err != nil {
			return nil, err
		}
		var result database.Result
		var err *cd.Error
		switch {
		case create && step.field == nil:
			result, err = s.sqlBuilder.BuildCreateTable(step.model)
		case create:
			result, err = s.sqlBuilder.BuildCreateRelationTable(step.model, step.field)
		case step.field == nil:
			result, err = s.sqlBuilder.BuildDropTable(step.model)
		default:
			result, err = s.sqlBuilder.BuildDropRelationTable(step.model, step.field)
		}
		if err != nil {
			return nil, err
		}
		statements = append(statements, result)
	}
	return statements, nil
}

func (s *baseRunner) executeSchemaDDL(statements []database.Result) *cd.Error {
	for _, statement := range statements {
		if err := s.checkContext(); err != nil {
			return err
		}
		if _, err := s.executor.Execute(statement.SQL(), statement.Args()...); err != nil {
			return err
		}
	}
	return s.checkContext()
}
