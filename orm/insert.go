package orm

import (
	"context"
	"fmt"
	"time"

	cd "github.com/muidea/magicCommon/def"

	"log/slog"

	"github.com/muidea/magicOrm/internal/database"
	"github.com/muidea/magicOrm/internal/database/codec"
	"github.com/muidea/magicOrm/internal/metrics/shared"
	"github.com/muidea/magicOrm/internal/utils"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/validation/errors"
)

func isRequiredRelationField(field models.Field) bool {
	if field == nil || field.GetSpec() == nil {
		return false
	}

	constraints := field.GetSpec().GetConstraints()
	return constraints != nil && constraints.Has(models.KeyRequired)
}

type insertRunner struct {
	baseRunner
	queryRunner
}

func newInsertRunner(
	ctx context.Context,
	vModel models.Model,
	executor database.Executor,
	provider provider.Provider,
	modelCodec codec.Codec) *insertRunner {
	baseRunner := newBaseRunner(ctx, vModel, executor, provider, modelCodec, false, 0)
	return &insertRunner{
		baseRunner: baseRunner,
		queryRunner: queryRunner{
			baseRunner: baseRunner,
		},
	}
}

func (s *insertRunner) insertHost(vModel models.Model) (err *cd.Error) {
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
		slog.Error("insertRunner insertHost innerHost failed", "error", pkErr.Error())
		return
	}

	if pkVal != nil && autoIncrementFlag {
		pkField := vModel.GetPrimaryField()
		vVal, vErr := s.modelCodec.ExtractBasicFieldValue(pkField, pkVal)
		if vErr != nil {
			err = vErr
			slog.Error("insertRunner insertHost ExtractBasicFieldValue/SetValue failed", "field", pkField.GetName(), "error", err.Error())
			return
		}
		err = pkField.SetValue(vVal)
		if err != nil {
			slog.Error("insertRunner insertHost ExtractBasicFieldValue/SetValue failed", "field", pkField.GetName(), "error", err.Error())
			return
		}
	}
	return
}

func (s *insertRunner) innerHost(vModel models.Model) (ret any, err *cd.Error) {
	insertResult, insertErr := s.sqlBuilder.BuildInsert(vModel)
	if insertErr != nil {
		err = insertErr
		slog.Error("insertRunner insertHost failed", "error", err.Error())
		return
	}

	var idVal any
	idErr := s.executor.ExecuteInsert(insertResult.SQL(), &idVal, insertResult.Args()...)
	if idErr != nil {
		err = idErr
		slog.Error("insertRunner insertHost failed", "error", err.Error())
		return
	}

	ret = idVal
	return
}

func (s *insertRunner) insertRelation(vModel models.Model, vField models.Field) (err *cd.Error) {
	if models.IsSliceField(vField) {
		rErr := s.insertSliceRelation(vModel, vField)
		if rErr != nil {
			err = rErr
			slog.Error("insertRunner failed", "error", err.Error())
			return
		}
		return
	}

	rErr := s.insertSingleRelation(vModel, vField)
	if rErr != nil {
		err = rErr
		slog.Error("insertRunner failed", "error", err.Error())
		return
	}
	return
}

func (s *insertRunner) insertSingleRelation(vModel models.Model, vField models.Field) (err *cd.Error) {
	elemType := vField.GetType().Elem()
	rModel, rErr := s.modelProvider.GetTypeModel(elemType)
	if rErr != nil {
		err = rErr
		slog.Error("insertRunner insertHost failed", "error", err.Error())
		return
	}
	rModel, rErr = s.modelProvider.SetModelValue(rModel, vField.GetValue())
	if rErr != nil {
		err = rErr
		slog.Error("insertRunner insertHost failed", "error", err.Error())
		return
	}

	if !models.IsPtrField(vField) {
		rInsertRunner := newInsertRunner(s.context, rModel, s.executor, s.modelProvider, s.modelCodec)
		rModel, rErr = rInsertRunner.Insert()
		if rErr != nil {
			err = rErr
			slog.Error("insertRunner insertRelation failed", "pkgKey", vModel.GetPkgKey(), "error", rErr.Error())
			return
		}
	}

	relationSQL, relationErr := s.sqlBuilder.BuildInsertRelation(vModel, vField, rModel)
	if relationErr != nil {
		err = relationErr
		slog.Error("insertRunner insertHost failed", "error", err.Error())
		return
	}

	var idVal any
	err = s.executor.ExecuteInsert(relationSQL.SQL(), &idVal, relationSQL.Args()...)
	if err != nil {
		slog.Error("insertRunner insertHost failed", "error", err.Error())
		return
	}

	vField.SetValue(rModel.Interface(models.IsPtrField(vField)))
	return
}

func (s *insertRunner) insertSliceRelation(vModel models.Model, vField models.Field) (err *cd.Error) {
	fSliceValue := vField.GetSliceValue()
	for _, fVal := range fSliceValue {
		elemType := vField.GetType().Elem()
		rModel, rErr := s.modelProvider.GetTypeModel(elemType)
		if rErr != nil {
			err = rErr
			slog.Error("insertRunner insertRelation failed", "field", vField.GetName(), "error", err.Error())
			return
		}
		rModel, rErr = s.modelProvider.SetModelValue(rModel, fVal)
		if rErr != nil {
			err = rErr
			slog.Error("insertRunner insertRelation failed", "field", vField.GetName(), "error", err.Error())
			return
		}

		if !elemType.IsPtrType() {
			rInsertRunner := newInsertRunner(s.context, rModel, s.executor, s.modelProvider, s.modelCodec)
			rModel, rErr = rInsertRunner.Insert()
			if rErr != nil {
				err = rErr
				slog.Error("insertRunner insertRelation failed", "field", vField.GetName(), "error", err.Error())
				return
			}
		}

		relationResult, relationErr := s.sqlBuilder.BuildInsertRelation(vModel, vField, rModel)
		if relationErr != nil {
			err = relationErr
			slog.Error("insertRunner insertRelation failed", "field", vField.GetName(), "error", err.Error())
			return
		}

		var idVal any
		err = s.executor.ExecuteInsert(relationResult.SQL(), &idVal, relationResult.Args()...)
		if err != nil {
			slog.Error("insertRunner insertRelation failed", "field", vField.GetName(), "error", err.Error())
			return
		}

		// 这里只需要直接更新值就可以
		err = fVal.Set(rModel.Interface(elemType.IsPtrType()))
		if err != nil {
			slog.Error("insertRunner insertRelation failed", "field", vField.GetName(), "error", err.Error())
			return
		}
	}
	return
}

// 引用只保存已有目标的主键；不能把空对象当作待分配 ID 的包含对象。
func (s *insertRunner) validateReferenceKeys(field models.Field) *cd.Error {
	elemType := field.GetType().Elem()
	if !elemType.IsPtrType() {
		return nil
	}
	values := []models.Value{field.GetValue()}
	if models.IsSliceField(field) {
		values = field.GetSliceValue()
	}
	for _, value := range values {
		target, err := s.modelProvider.GetTypeModel(elemType)
		if err != nil {
			return err
		}
		target, err = s.modelProvider.SetModelValue(target, value)
		if err != nil {
			return err
		}
		primary := target.GetPrimaryField()
		if primary == nil || primary.GetValue() == nil || !primary.GetValue().IsValid() || primary.GetValue().IsZero() {
			return cd.NewError(cd.IllegalParam, fmt.Sprintf("reference relation field %s has entity without primary key", field.GetName()))
		}
	}
	return nil
}

func (s *insertRunner) Insert() (ret models.Model, err *cd.Error) {
	if err = s.checkContext(); err != nil {
		return
	}

	// Validate relation presence before writing the host. Assignment records
	// explicit clears for updates; an assigned nil optional reference still has
	// no relation to insert. Do not change that shared assignment state.
	relations := models.Fields{}
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
			slog.Error("insertRunner Insert illegal field value", "field", field.GetName(), "error", err.Error())
			return
		}
		if models.IsStructField(field) && !models.IsValidField(field) {
			if models.IsPtrField(field) && !isRequiredRelationField(field) {
				continue
			}
			return nil, cd.NewError(cd.IllegalParam, fmt.Sprintf("illegal field value, field:%s", field.GetName()))
		}
		if err = s.validateReferenceKeys(field); err != nil {
			return nil, err
		}
		if !models.IsInlineReference(s.vModel, field) {
			relations = append(relations, field)
		}
	}

	err = s.insertHost(s.vModel)
	if err != nil {
		slog.Error("insertRunner insertHost failed", "error", err.Error())
		return
	}

	for _, field := range relations {
		err = s.insertRelation(s.vModel, field)
		if err != nil {
			slog.Error("insertRunner failed", "error", err.Error())
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

	if err = s.checkContext(); err != nil {
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
		if recovered := recover(); recovered != nil {
			// A panic must never look like a successful operation to the deferred
			// transaction finalizer. Roll back and preserve the original panic.
			_ = s.finalTransaction(cd.NewError(cd.Unexpected, "insert aborted by panic"))
			panic(recovered)
		}
		if transactionErr := s.finalTransaction(err); err == nil && transactionErr != nil {
			ret = nil
			err = transactionErr
		}
	}()

	insertRunner := newInsertRunner(s.context, vModel, s.executor, s.modelProvider, s.modelCodec)
	ret, err = insertRunner.Insert()
	if err != nil {
		slog.Error("Insert insertRunner.Insert failed", "pkgKey", vModel.GetPkgKey(), "error", err.Error())
		return
	}
	return
}
