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

type DropRunner struct {
	baseRunner
}

func NewDropRunner(ctx context.Context, vModel models.Model, executor database.Executor, provider provider.Provider, modelCodec codec.Codec) *DropRunner {
	return &DropRunner{
		baseRunner: newBaseRunner(ctx, vModel, executor, provider, modelCodec, false, 0),
	}
}

func (s *DropRunner) dropHost(vModel models.Model) (err *cd.Error) {
	dropResult, dropErr := s.sqlBuilder.BuildDropTable(vModel)
	if dropErr != nil {
		err = dropErr
		slog.Error("operation failed", "error", "operation failed")
		return
	}

	_, err = s.executor.Execute(dropResult.SQL(), dropResult.Args()...)
	if err != nil {
		slog.Error("operation failed", "error", "operation failed")
	}
	return
}

func (s *DropRunner) dropRelation(vModel models.Model, vField models.Field) (err *cd.Error) {
	relationResult, relationErr := s.sqlBuilder.BuildDropRelationTable(vModel, vField)
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

func (s *DropRunner) Drop() (err *cd.Error) {
	err = s.dropHost(s.vModel)
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

			rRunner := NewDropRunner(s.context, rModel, s.executor, s.modelProvider, s.modelCodec)
			err = rRunner.Drop()
			if err != nil {
				slog.Error("operation failed", "error", err.Error())
				return
			}
		}

		err = s.dropRelation(s.vModel, field)
		if err != nil {
			slog.Error("operation failed", "error", err.Error())
			return
		}
	}

	return
}

func (s *impl) Drop(vModel models.Model) (err *cd.Error) {
	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime)
		if ormMetricCollector != nil {
			ormMetricCollector.RecordOperation("drop", vModel, duration, err)
		}
	}()

	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}

	dropRunner := NewDropRunner(s.context, vModel, s.executor, s.modelProvider, s.modelCodec)
	err = dropRunner.Drop()
	if err != nil {
		slog.Error("operation failed", "error", "operation failed")
	}
	return
}
