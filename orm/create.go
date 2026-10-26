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

type createRunner struct {
	baseRunner
}

func newCreateRunner(ctx context.Context, vModel models.Model, executor database.Executor, provider provider.Provider, modelCodec codec.Codec) *createRunner {
	return &createRunner{
		baseRunner: newBaseRunner(ctx, vModel, executor, provider, modelCodec, false, 0),
	}
}

func (s *createRunner) Create() *cd.Error {
	graph, err := collectSchemaGraph(s.context, s.vModel, s.modelProvider, s.modelCodec)
	if err != nil {
		return err
	}
	statements, err := s.buildSchemaDDL(graph, true)
	if err != nil {
		return err
	}
	return s.executeSchemaDDL(statements)
}

func (s *impl) Create(vModel models.Model) (err *cd.Error) {
	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime)
		if ormMetricCollector != nil {
			ormMetricCollector.RecordOperation(string(metrics.OperationCreate), vModel, duration, cd.ToStdError(err))
		}
	}()

	if err = s.checkContext(); err != nil {
		return
	}

	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}

	createRunner := newCreateRunner(s.context, vModel, s.executor, s.modelProvider, s.modelCodec)
	err = createRunner.Create()
	if err != nil {
		slog.Error("Create createRunner.Create failed", "pkgKey", vModel.GetPkgKey(), "error", err.Error())
	}
	return
}
