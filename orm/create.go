package orm

import (
	"context"
	"time"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"log/slog"
)

type CreateRunner struct {
	baseRunner
}

func NewCreateRunner(ctx context.Context, vModel models.Model, executor database.Executor, provider provider.Provider, modelCodec codec.Codec) *CreateRunner {
	return &CreateRunner{
		baseRunner: newBaseRunner(ctx, vModel, executor, provider, modelCodec, false, 0),
	}
}

func (s *CreateRunner) createHost() (err *cd.Error) {
	createResult, createErr := s.sqlBuilder.BuildCreateTable(s.vModel)
	if createErr != nil {
		err = createErr
		slog.Error("operation failed", "error", "operation failed")
		return
	}

	_, err = s.executor.Execute(createResult.SQL(), createResult.Args()...)
	if err != nil {
		slog.Error("operation failed", "error", "operation failed")
	}
	return
}

func (s *CreateRunner) createRelation(vField models.Field) (err *cd.Error) {
	relationResult, relationErr := s.sqlBuilder.BuildCreateRelationTable(s.vModel, vField)
	if relationErr != nil {
		err = relationErr
		slog.Error("operation failed", "error", "operation failed")
		return
	}

	_, err = s.executor.Execute(relationResult.SQL(), relationResult.Args()...)
	if err != nil {
		slog.Error("operation failed", "error", "operation failed")
	}
	return
}

func (s *CreateRunner) Create() (err *cd.Error) {
	if err = s.checkContext(); err != nil {
		return
	}

	err = s.createHost()
	if err != nil {
		slog.Error("operation failed", "error", "operation failed")
		return
	}

	for _, field := range s.vModel.GetFields() {
		if models.IsBasicField(field) {
			continue
		}

		elemType := field.GetType().Elem()
		if !elemType.IsPtrType() {
			rModel, rErr := s.modelProvider.GetTypeModel(elemType)
			if rErr != nil {
				err = rErr
				slog.Error("operation failed", "error", err.Error())
				return
			}

			rRunner := NewCreateRunner(s.context, rModel, s.executor, s.modelProvider, s.modelCodec)
			err = rRunner.Create()
			if err != nil {
				slog.Error("operation failed", "error", err.Error())
				return
			}
		}

		err = s.createRelation(field)
		if err != nil {
			slog.Error("operation failed", "error", err.Error())
			return
		}
	}

	return
}

func (s *impl) Create(vModel models.Model) (err *cd.Error) {
	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime)
		if ormMetricCollector != nil {
			ormMetricCollector.RecordOperation("create", vModel, duration, err)
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
		slog.Error("operation failed", "error", "operation failed")
	}
	return
}
