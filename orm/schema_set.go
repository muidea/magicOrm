package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/internal/database"
	"github.com/muidea/magicOrm/models"
)

var _ models.SchemaSetLifecycle = (*impl)(nil)

func (s *impl) DropSchemaSet(roots []models.Model) (*SchemaReport, *cd.Error) {
	if err := s.checkContext(); err != nil {
		return nil, err
	}
	reader, ok := s.executor.(database.SchemaReader)
	if !ok {
		return nil, cd.NewError(cd.Unexpected, "executor does not support physical schema inspection")
	}
	graph, err := collectSchemaSetGraph(s.context, roots, s.modelProvider, s.modelCodec)
	if err != nil {
		return nil, err
	}
	// Remove all relationship tables before their hosts. No host or shared
	// owned table is handed to another independent Drop invocation.
	order := make([]schemaTable, 0, len(graph.order))
	for i := len(graph.order) - 1; i >= 0; i-- {
		if graph.order[i].field != nil {
			order = append(order, graph.order[i])
		}
	}
	for i := len(graph.order) - 1; i >= 0; i-- {
		if graph.order[i].field == nil {
			order = append(order, graph.order[i])
		}
	}
	graph.order = order
	runner := newBaseRunner(s.context, roots[0], s.executor, s.modelProvider, s.modelCodec, false, 0)
	// Build the entire declaration before inspecting/applying the repair plan.
	if _, err := runner.buildSchemaDDL(graph, false); err != nil {
		return nil, err
	}
	statements, err := runner.planSchemaRepair(graph, nil, reader, true)
	if err != nil {
		return nil, err
	}
	if err := runner.executeSchemaDDL(statements); err != nil {
		return nil, err
	}
	report, err := s.InspectSchemaSet(roots, false)
	if err != nil {
		return nil, err
	}
	if !report.Matches {
		return nil, cd.NewError(cd.DataCorrupted, "schema set drop physical confirmation failed")
	}
	return report, nil
}
