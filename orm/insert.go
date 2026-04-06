package orm

import (
	"context"
	"fmt"
	"time"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/metrics"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/utils"
	"github.com/muidea/magicOrm/validation/errors"
	"log/slog"
)

func isRequiredRelationField(field models.Field) bool {
	if field == nil || field.GetSpec() == nil {
		return false
	}

	constraints := field.GetSpec().GetConstraints()
	return constraints != nil && constraints.Has(models.KeyRequired)
}

type InsertRunner struct {
	baseRunner
	QueryRunner
}

func NewInsertRunner(
	ctx context.Context,
	vModel models.Model,
	executor database.Executor,
	provider provider.Provider,
	modelCodec codec.Codec) *InsertRunner {
	baseRunner := newBaseRunner(ctx, vModel, executor, provider, modelCodec, false, 0)
	return &InsertRunner{
		baseRunner: baseRunner,
		QueryRunner: QueryRunner{
			baseRunner: baseRunner,
		},
	}
}

func (s *InsertRunner) insertHost(vModel models.Model) (err *cd.Error) {
	autoIncrementFlag := false
	for _, field := range vModel.GetFields() {
		if !models.IsBasicField(field) {
			continue
		}

		vVal := field.GetValue()
		switch field.GetSpec().GetValueDeclare() {
		case models.AutoIncrement:
			autoIncrementFlag = true
		case models.UUID:
			if vVal.IsZero() {
				if setErr := vVal.Set(utils.GetNewUUID()); setErr != nil {
					slog.Warn("Failed to set UUID value", "error", setErr)
				}
			}
		case models.Snowflake:
			if vVal.IsZero() {
				if setErr := vVal.Set(utils.GetNewSnowflakeID()); setErr != nil {
					slog.Warn("Failed to set snowflake ID", "error", setErr)
				}
			}
		case models.DateTime:
			if vVal.IsZero() {
				if setErr := vVal.Set(utils.GetCurrentDateTime()); setErr != nil {
					slog.Warn("Failed to set datetime", "error", setErr)
				}
			}
		}
	}

	pkVal, pkErr := s.innerHost(vModel)
	if pkErr != nil {
		err = pkErr
		slog.Error("InsertRunner insertHost innerHost failed", "error", pkErr.Error())
		return
	}

	if pkVal != nil && autoIncrementFlag {
		pkField := vModel.GetPrimaryField()
		vVal, vErr := s.modelCodec.ExtractBasicFieldValue(pkField, pkVal)
		if vErr != nil {
			err = vErr
			slog.Error("InsertRunner insertHost ExtractBasicFieldValue/SetValue failed", "field", pkField.GetName(), "error", err.Error())
			return
		}
		err = pkField.SetValue(vVal)
		if err != nil {
			slog.Error("InsertRunner insertHost ExtractBasicFieldValue/SetValue failed", "field", pkField.GetName(), "error", err.Error())
			return
		}
	}
	return
}

func (s *InsertRunner) innerHost(vModel models.Model) (ret any, err *cd.Error) {
	insertResult, insertErr := s.sqlBuilder.BuildInsert(vModel)
	if insertErr != nil {
		err = insertErr
		slog.Error("InsertRunner insertHost failed", "error", err.Error())
		return
	}

	var idVal any
	idErr := s.executor.ExecuteInsert(insertResult.SQL(), &idVal, insertResult.Args()...)
	if idErr != nil {
		err = idErr
		slog.Error("InsertRunner insertHost failed", "error", err.Error())
		return
	}

	ret = idVal
	return
}

func (s *InsertRunner) insertRelation(vModel models.Model, vField models.Field) (err *cd.Error) {
	if models.IsSliceField(vField) {
		rErr := s.insertSliceRelation(vModel, vField)
		if rErr != nil {
			err = rErr
			slog.Error("InsertRunner failed", "error", err.Error())
			return
		}
		return
	}

	rErr := s.insertSingleRelation(vModel, vField)
	if rErr != nil {
		err = rErr
		slog.Error("InsertRunner failed", "error", err.Error())
		return
	}
	return
}

func (s *InsertRunner) insertSingleRelation(vModel models.Model, vField models.Field) (err *cd.Error) {
	elemType := vField.GetType().Elem()
	rModel, rErr := s.modelProvider.GetTypeModel(elemType)
	if rErr != nil {
		err = rErr
		slog.Error("InsertRunner insertHost failed", "error", err.Error())
		return
	}
	rModel, rErr = s.modelProvider.SetModelValue(rModel, vField.GetValue())
	if rErr != nil {
		err = rErr
		slog.Error("InsertRunner insertHost failed", "error", err.Error())
		return
	}

	if !models.IsPtrField(vField) {
		rInsertRunner := NewInsertRunner(s.context, rModel, s.executor, s.modelProvider, s.modelCodec)
		rModel, rErr = rInsertRunner.Insert()
		if rErr != nil {
			err = rErr
			slog.Error("InsertRunner insertRelation failed", "pkgKey", vModel.GetPkgKey(), "error", rErr.Error())
			return
		}
	}

	relationSQL, relationErr := s.sqlBuilder.BuildInsertRelation(vModel, vField, rModel)
	if relationErr != nil {
		err = relationErr
		slog.Error("InsertRunner insertHost failed", "error", err.Error())
		return
	}

	var idVal any
	err = s.executor.ExecuteInsert(relationSQL.SQL(), &idVal, relationSQL.Args()...)
	if err != nil {
		slog.Error("InsertRunner insertHost failed", "error", err.Error())
		return
	}

	vField.SetValue(rModel.Interface(models.IsPtrField(vField)))
	return
}

func (s *InsertRunner) insertSliceRelation(vModel models.Model, vField models.Field) (err *cd.Error) {
	fSliceValue := vField.GetSliceValue()
	for _, fVal := range fSliceValue {
		elemType := vField.GetType().Elem()
		rModel, rErr := s.modelProvider.GetTypeModel(elemType)
		if rErr != nil {
			err = rErr
			slog.Error("InsertRunner insertRelation failed", "field", vField.GetName(), "error", err.Error())
			return
		}
		rModel, rErr = s.modelProvider.SetModelValue(rModel, fVal)
		if rErr != nil {
			err = rErr
			slog.Error("InsertRunner insertRelation failed", "field", vField.GetName(), "error", err.Error())
			return
		}

		if !elemType.IsPtrType() {
			rInsertRunner := NewInsertRunner(s.context, rModel, s.executor, s.modelProvider, s.modelCodec)
			rModel, rErr = rInsertRunner.Insert()
			if rErr != nil {
				err = rErr
				slog.Error("InsertRunner insertRelation failed", "field", vField.GetName(), "error", err.Error())
				return
			}
		}

		relationResult, relationErr := s.sqlBuilder.BuildInsertRelation(vModel, vField, rModel)
		if relationErr != nil {
			err = relationErr
			slog.Error("InsertRunner insertRelation failed", "field", vField.GetName(), "error", err.Error())
			return
		}

		var idVal any
		err = s.executor.ExecuteInsert(relationResult.SQL(), &idVal, relationResult.Args()...)
		if err != nil {
			slog.Error("InsertRunner insertRelation failed", "field", vField.GetName(), "error", err.Error())
			return
		}

		// 这里只需要直接更新值就可以
		err = fVal.Set(rModel.Interface(elemType.IsPtrType()))
		if err != nil {
			slog.Error("InsertRunner insertRelation failed", "field", vField.GetName(), "error", err.Error())
			return
		}
	}
	return
}

func (s *InsertRunner) Insert() (ret models.Model, err *cd.Error) {
	if err = s.checkContext(); err != nil {
		return
	}

	err = s.insertHost(s.vModel)
	if err != nil {
		slog.Error("InsertRunner insertHost failed", "error", err.Error())
		return
	}

	for _, field := range s.vModel.GetFields() {
		// 忽略基础字段
		if models.IsBasicField(field) {
			continue
		}

		if !models.IsAssignedField(field) {
			// 未赋值的关系字段默认可跳过，但 required relation 仍然必须提供。
			if (field.GetType().IsPtrType() || models.IsSliceField(field)) && !isRequiredRelationField(field) {
				continue
			}

			err = cd.NewError(cd.IllegalParam, fmt.Sprintf("illegal field value, field:%s", field.GetName()))
			slog.Error("InsertRunner Insert illegal field value", "field", field.GetName(), "error", err.Error())
			return
		}

		err = s.insertRelation(s.vModel, field)
		if err != nil {
			slog.Error("InsertRunner failed", "error", err.Error())
			return
		}
	}

	ret, err = projectWriteResponseModel(s.vModel, s.modelProvider)
	return
}

func (s *impl) Insert(vModel models.Model) (ret models.Model, err *cd.Error) {
	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime)
		if ormMetricCollector != nil {
			ormMetricCollector.RecordOperation(string(metrics.OperationInsert), vModel, duration, cd.ToStdError(err))
		}
	}()

	if err = s.CheckContext(); err != nil {
		return
	}

	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}

	// Validate model before insertion
	validationErr := s.validateModel(vModel, errors.ScenarioInsert)
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

	insertRunner := NewInsertRunner(s.context, vModel, s.executor, s.modelProvider, s.modelCodec)
	ret, err = insertRunner.Insert()
	if err != nil {
		slog.Error("Insert InsertRunner.Insert failed", "pkgKey", vModel.GetPkgKey(), "error", err.Error())
		return
	}
	return
}
