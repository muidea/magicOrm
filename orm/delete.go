package orm

import (
	"context"
	"time"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/metrics"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/validation/errors"
	"log/slog"
)

type DeleteRunner struct {
	baseRunner
	QueryRunner
}

func NewDeleteRunner(
	ctx context.Context,
	vModel models.Model,
	executor database.Executor,
	provider provider.Provider,
	modelCodec codec.Codec,
	deepLevel int) *DeleteRunner {
	baseRunner := newBaseRunner(ctx, vModel, executor, provider, modelCodec, false, deepLevel)
	return &DeleteRunner{
		baseRunner: baseRunner,
		QueryRunner: QueryRunner{
			baseRunner: baseRunner,
		},
	}
}

func (s *DeleteRunner) deleteHost(vModel models.Model) (err *cd.Error) {
	deleteResult, deleteErr := s.sqlBuilder.BuildDelete(vModel)
	if deleteErr != nil {
		err = deleteErr
		slog.Error("DeleteRunner deleteHost BuildDelete failed", "error", err.Error())
		return
	}

	_, err = s.executor.Execute(deleteResult.SQL(), deleteResult.Args()...)
	if err != nil {
		slog.Error("DeleteRunner deleteHost Execute failed", "error", err.Error())
	}
	return
}

func (s *DeleteRunner) deleteRelation(vModel models.Model, vField models.Field, deepLevel int) (err *cd.Error) {
	hostResult, relationResult, resultErr := s.sqlBuilder.BuildDeleteRelation(vModel, vField)
	if resultErr != nil {
		err = resultErr
		slog.Error("DeleteRunner failed", "error", err.Error())
		return
	}

	vType := vField.GetType()
	// 这里使用ElemType 是因为vFiled的指针表示该字段是否可选，
	elemType := vType.Elem()
	if !elemType.IsPtrType() {
		fieldErr := s.queryRelation(vModel, vField, maxDeepLevel-1)
		if fieldErr != nil {
			err = fieldErr
			slog.Error("DeleteRunner failed", "error", err.Error())
			return
		}

		if models.IsStructType(vType.GetValue()) {
			err = s.deleteRelationSingleStructInner(vField, deepLevel)
			if err != nil {
				slog.Error("DeleteRunner failed", "error", err.Error())
				return
			}
		} else if models.IsSliceType(vType.GetValue()) {
			err = s.deleteRelationSliceStructInner(vField, deepLevel)
			if err != nil {
				slog.Error("DeleteRunner failed", "error", err.Error())
				return
			}
		}

		_, err = s.executor.Execute(hostResult.SQL(), hostResult.Args()...)
		if err != nil {
			slog.Error("DeleteRunner failed", "error", err.Error())
			return
		}
	}

	_, err = s.executor.Execute(relationResult.SQL(), relationResult.Args()...)
	if err != nil {
		slog.Error("DeleteRunner failed", "error", err.Error())
	}
	return
}

func (s *DeleteRunner) deleteRelationSingleStructInner(vField models.Field, deepLevel int) (err *cd.Error) {
	rModel, rErr := s.modelProvider.GetTypeModel(vField.GetType())
	if rErr != nil {
		err = rErr
		slog.Error("DeleteRunner failed", "error", err.Error())
		return
	}

	rModel, rErr = s.modelProvider.SetModelValue(rModel, vField.GetValue())
	if rErr != nil {
		err = rErr
		slog.Error("DeleteRunner failed", "error", err.Error())
		return
	}

	rRunner := NewDeleteRunner(s.context, rModel, s.executor, s.modelProvider, s.modelCodec, deepLevel+1)
	err = rRunner.Delete()
	if err != nil {
		slog.Error("DeleteRunner failed", "error", err.Error())
		return
	}

	return
}

func (s *DeleteRunner) deleteRelationSliceStructInner(vField models.Field, deepLevel int) (err *cd.Error) {
	sliceVal := vField.GetSliceValue()
	for _, val := range sliceVal {
		rModel, rErr := s.modelProvider.GetTypeModel(vField.GetType().Elem())
		if rErr != nil {
			err = rErr
			slog.Error("DeleteRunner failed", "error", err.Error())
			return
		}

		rModel, rErr = s.modelProvider.SetModelValue(rModel, val)
		if rErr != nil {
			err = rErr
			slog.Error("DeleteRunner failed", "error", err.Error())
			return
		}
		rRunner := NewDeleteRunner(s.context, rModel, s.executor, s.modelProvider, s.modelCodec, deepLevel+1)
		err = rRunner.Delete()
		if err != nil {
			slog.Error("DeleteRunner failed", "error", err.Error())
			return
		}
	}

	return
}

func (s *DeleteRunner) Delete() (err *cd.Error) {
	if err = s.checkContext(); err != nil {
		return
	}

	err = s.deleteHost(s.vModel)
	if err != nil {
		slog.Error("DeleteRunner failed", "error", err.Error())
		return
	}

	for _, field := range s.vModel.GetFields() {
		if models.IsBasicField(field) {
			continue
		}

		err = s.deleteRelation(s.vModel, field, 0)
		if err != nil {
			slog.Error("DeleteRunner failed", "error", err.Error())
			return
		}
	}

	return
}

func (s *impl) Delete(vModel models.Model) (ret models.Model, err *cd.Error) {
	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime)
		if ormMetricCollector != nil {
			ormMetricCollector.RecordOperation(string(metrics.OperationDelete), vModel, duration, err)
		}
	}()

	if err = s.CheckContext(); err != nil {
		return
	}

	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}

	// Validate model before deletion (minimal validation for delete scenario)
	validationErr := s.validateModel(vModel, errors.ScenarioDelete)
	if validationErr != nil {
		err = validationErr
		return
	}

	err = s.executor.BeginTransaction()
	if err != nil {
		return
	}
	defer func() {
		s.finalTransaction(err)
	}()

	deleteRunner := NewDeleteRunner(s.context, vModel, s.executor, s.modelProvider, s.modelCodec, 0)
	err = deleteRunner.Delete()
	if err != nil {
		slog.Error("Delete DeleteRunner.Delete failed", "pkgKey", vModel.GetPkgKey(), "error", err.Error())
		return
	}

	ret = vModel
	return
}
