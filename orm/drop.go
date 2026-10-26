package orm

import (
	"context"
	"time"

	cd "github.com/muidea/magicCommon/def"

	"log/slog"

	"github.com/muidea/magicOrm/internal/database"
	"github.com/muidea/magicOrm/internal/database/codec"
	"github.com/muidea/magicOrm/internal/metrics/shared"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

type dropRunner struct {
	baseRunner
}

func newDropRunner(ctx context.Context, vModel models.Model, executor database.Executor, provider provider.Provider, modelCodec codec.Codec) *dropRunner {
	return &dropRunner{
		baseRunner: newBaseRunner(ctx, vModel, executor, provider, modelCodec, false, 0),
	}
}

func (s *dropRunner) Drop() *cd.Error {
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

	dropRunner := newDropRunner(s.context, vModel, s.executor, s.modelProvider, s.modelCodec)
	err = dropRunner.Drop()
	if err != nil {
		slog.Error("Drop dropRunner.Drop failed", "pkgKey", vModel.GetPkgKey(), "error", err.Error())
	}
	return
}
