package orm

import (
	"context"
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/utils"
)

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
				vVal.Set(utils.GetNewUUID())
			}
		case models.Snowflake:
			if vVal.IsZero() {
				vVal.Set(utils.GetNewSnowflakeID())
			}
		case models.DateTime:
			if vVal.IsZero() {
				vVal.Set(utils.GetCurrentDateTime())
			}
		}
	}

	pkVal, pkErr := s.innerHost(vModel)
	if pkErr != nil {
		err = pkErr
		log.Errorf("insertHost failed, s.innerHost error:%s", err.Error())
		return
	}

	if pkVal != nil && autoIncrementFlag {
		pkField := vModel.GetPrimaryField()
		vVal, vErr := s.modelCodec.ExtractBasicFieldValue(pkField, pkVal)
		if vErr != nil {
			err = vErr
			log.Errorf("insertHost failed, extract pkField:%s, pkField type:%s, s.modelCodec.ExtractFieldValue error:%s", pkField.GetName(), pkField.GetType().GetPkgKey(), err.Error())
			return
		}
		err = pkField.SetValue(vVal)
		if err != nil {
			log.Errorf("insertHost failed, set pkField:%s, pkField type:%s, s.modelCodec.ExtractFieldValue error:%s", pkField.GetName(), pkField.GetType().GetPkgKey(), err.Error())
			return
		}
	}
	return
}

func (s *InsertRunner) innerHost(vModel models.Model) (ret any, err *cd.Error) {
	insertResult, insertErr := s.sqlBuilder.BuildInsert(vModel)
	if insertErr != nil {
		err = insertErr
		log.Errorf("innerHost failed, builder.BuildInsert error:%s", err.Error())
		return
	}

	var idVal any
	idErr := s.executor.ExecuteInsert(insertResult.SQL(), &idVal, insertResult.Args()...)
	if idErr != nil {
		err = idErr
		log.Errorf("innerHost failed, s.executor.Execute error:%s", err.Error())
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
			log.Errorf("insertRelation failed, field:%s, s.insertSliceRelation error:%s", vField.GetName(), err.Error())
			return
		}
		return
	}

	rErr := s.insertSingleRelation(vModel, vField)
	if rErr != nil {
		err = rErr
		log.Errorf("insertRelation failed, field:%s, s.insertSingleRelation error:%s", vField.GetName(), err.Error())
		return
	}
	return
}

func (s *InsertRunner) insertSingleRelation(vModel models.Model, vField models.Field) (err *cd.Error) {
	elemType := vField.GetType().Elem()
	rModel, rErr := s.modelProvider.GetTypeModel(elemType)
	if rErr != nil {
		err = rErr
		log.Errorf("insertSingleRelation failed, s.modelProvider.GetTypeModel error:%s", err.Error())
		return
	}
	rModel, rErr = s.modelProvider.SetModelValue(rModel, vField.GetValue())
	if rErr != nil {
		err = rErr
		log.Errorf("insertSingleRelation failed, s.modelProvider.SetModelValue error:%s", err.Error())
		return
	}

	if !models.IsPtrField(vField) {
		rInsertRunner := NewInsertRunner(s.context, rModel, s.executor, s.modelProvider, s.modelCodec)
		rModel, rErr = rInsertRunner.Insert()
		if rErr != nil {
			err = rErr
			log.Errorf("insertSingleRelation failed, rInsertRunner.Insert() error:%s", err.Error())
			return
		}
	}

	relationSQL, relationErr := s.sqlBuilder.BuildInsertRelation(vModel, vField, rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("insertSingleRelation failed, s.sqlBuilder.BuildInsertRelation error:%s", err.Error())
		return
	}

	var idVal any
	err = s.executor.ExecuteInsert(relationSQL.SQL(), &idVal, relationSQL.Args()...)
	if err != nil {
		log.Errorf("insertSingleRelation failed, s.executor.Execute error:%s", err.Error())
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
			log.Errorf("insertSliceRelation failed, model:%s, filed name:%s, s.modelProvider.GetTypeModel error:%s", vModel.GetPkgKey(), vField.GetName(), err.Error())
			return
		}
		rModel, rErr = s.modelProvider.SetModelValue(rModel, fVal)
		if rErr != nil {
			err = rErr
			log.Errorf("insertSliceRelation failed, model:%s, filed name:%s, s.modelProvider.SetModelValue error:%s", vModel.GetPkgKey(), vField.GetName(), err.Error())
			return
		}

		if !elemType.IsPtrType() {
			rInsertRunner := NewInsertRunner(s.context, rModel, s.executor, s.modelProvider, s.modelCodec)
			rModel, rErr = rInsertRunner.Insert()
			if rErr != nil {
				err = rErr
				log.Errorf("insertSliceRelation failed, model:%s, filed name:%s, s.insertSingle error:%s", vModel.GetPkgKey(), vField.GetName(), err.Error())
				return
			}
		}

		relationResult, relationErr := s.sqlBuilder.BuildInsertRelation(vModel, vField, rModel)
		if relationErr != nil {
			err = relationErr
			log.Errorf("insertSliceRelation failed, model:%s, filed name:%s, s.sqlBuilder.BuildInsertRelation error:%s", vModel.GetPkgKey(), vField.GetName(), err.Error())
			return
		}

		var idVal any
		err = s.executor.ExecuteInsert(relationResult.SQL(), &idVal, relationResult.Args()...)
		if err != nil {
			log.Errorf("insertSliceRelation failed, model:%s, filed name:%s, s.executor.Execute error:%s", vModel.GetPkgKey(), vField.GetName(), err.Error())
			return
		}

		// 这里只需要直接更新值就可以
		err = fVal.Set(rModel.Interface(elemType.IsPtrType()))
		if err != nil {
			log.Errorf("insertSliceRelation failed, model:%s, filed name:%s, fVal.Set error:%s", vModel.GetPkgKey(), vField.GetName(), err.Error())
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
		log.Errorf("Insert failed, s.insertSingle error:%s", err.Error())
		return
	}

	for _, field := range s.vModel.GetFields() {
		// 忽略基础字段
		if models.IsBasicField(field) {
			continue
		}

		if !models.IsAssignedField(field) {
			// 未赋值, 如果是可选字段，或者是空slice字段，则忽略
			if field.GetType().IsPtrType() || models.IsSliceField(field) {
				continue
			}

			// 未赋值，但是是必选字段，则需要报错提示
			err = cd.NewError(cd.IllegalParam, fmt.Sprintf("illegal field value, field:%s", field.GetName()))
			log.Errorf("Insert field:%s model failed, s.insertRelation error:%s", field.GetName(), err.Error())
			return
		}

		err = s.insertRelation(s.vModel, field)
		if err != nil {
			log.Errorf("Insert relation field:%s failed, s.insertRelation error:%s", field.GetName(), err.Error())
			return
		}
	}

	ret = s.vModel
	return
}

func (s *impl) Insert(vModel models.Model) (ret models.Model, err *cd.Error) {
	if err = s.CheckContext(); err != nil {
		return
	}

	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}

	err = s.executor.BeginTransaction()
	if err != nil {
		return
	}
	defer s.finalTransaction(err)

	insertRunner := NewInsertRunner(s.context, vModel, s.executor, s.modelProvider, s.modelCodec)
	ret, err = insertRunner.Insert()
	if err != nil {
		log.Errorf("Insert failed, insertRunner.Insert() error:%s", err.Error())
		return
	}
	return
}
