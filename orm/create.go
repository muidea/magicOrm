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

type CreateRunner struct {
	baseRunner
}

func NewCreateRunner(ctx context.Context, vModel models.Model, executor database.Executor, provider provider.Provider, modelCodec codec.Codec) *CreateRunner {
	return &CreateRunner{
		baseRunner: newBaseRunner(ctx, vModel, executor, provider, modelCodec, false, 0),
	}
}

func (s *CreateRunner) Create() *cd.Error {
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

	if err = s.CheckContext(); err != nil {
		return
	}

	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}

	createRunner := NewCreateRunner(s.context, vModel, s.executor, s.modelProvider, s.modelCodec)
	err = createRunner.Create()
	if err != nil {
		slog.Error("Create CreateRunner.Create failed", "pkgKey", vModel.GetPkgKey(), "error", err.Error())
	}
	return
}
