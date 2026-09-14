package orm

import (
	"context"
	"time"

	cd "github.com/muidea/magicCommon/def"

	"log/slog"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/metrics"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

type DropRunner struct {
	baseRunner
}

func NewDropRunner(ctx context.Context, vModel models.Model, executor database.Executor, provider provider.Provider, modelCodec codec.Codec) *DropRunner {
	return &DropRunner{
		baseRunner: newBaseRunner(ctx, vModel, executor, provider, modelCodec, false, 0),
	}
}

func (s *DropRunner) Drop() *cd.Error {
	graph, err := collectSchemaGraph(s.context, s.vModel, s.modelProvider, s.modelCodec)
	if err != nil {
		return err
	}
	statements, err := s.buildSchemaDDL(graph, false)
	if err != nil {
		return err
	}
	return s.executeSchemaDDL(statements)
}

func (s *impl) Drop(vModel models.Model) (err *cd.Error) {
	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime)
		if ormMetricCollector != nil {
			ormMetricCollector.RecordOperation(string(metrics.OperationDrop), vModel, duration, cd.ToStdError(err))
		}
	}()

	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}

	dropRunner := NewDropRunner(s.context, vModel, s.executor, s.modelProvider, s.modelCodec)
	err = dropRunner.Drop()
	if err != nil {
		slog.Error("Drop DropRunner.Drop failed", "pkgKey", vModel.GetPkgKey(), "error", err.Error())
	}
	return
}
