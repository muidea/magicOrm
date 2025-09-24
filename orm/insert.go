package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/executor"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

type InsertRunner struct {
	baseRunner
	QueryRunner
}

func NewInsertRunner(
	vModel model.Model,
	executor executor.Executor,
	provider provider.Provider,
	modelCodec codec.Codec) *InsertRunner {
	baseRunner := newBaseRunner(vModel, executor, provider, modelCodec, false, 0)
	return &InsertRunner{
		baseRunner: baseRunner,
		QueryRunner: QueryRunner{
			baseRunner: baseRunner,
		},
	}
}

func (s *InsertRunner) insertHost(vModel model.Model) (err *cd.Error) {
	autoIncrementFlag := false
	for _, field := range vModel.GetFields() {
		if !model.IsBasicField(field) {
			continue
		}

		if field.GetSpec().GetValueDeclare() == model.AutoIncrement {
			autoIncrementFlag = true
		}
	}

	pkVal, pkErr := s.innerHost(vModel)
	if pkErr != nil {
		err = pkErr
		log.Errorf("insertHost failed, s.innerHost error:%s", err.Error())
		return
	}

	if pkVal != nil && autoIncrementFlag {
		pkFiled := vModel.GetPrimaryField()
		vVal, vErr := s.modelCodec.ExtractBasicFieldValue(pkFiled, pkVal)
		if vErr != nil {
			err = vErr
			log.Errorf("insertHost failed, s.modelCodec.ExtractFieldValue error:%s", err.Error())
			return
		}
		err = pkFiled.SetValue(vVal)
		if err != nil {
			log.Errorf("insertHost failed, s.modelCodec.ExtractFieldValue error:%s", err.Error())
			return
		}
	}
	return
}

func (s *InsertRunner) innerHost(vModel model.Model) (ret any, err *cd.Error) {
	insertResult, insertErr := s.hBuilder.BuildInsert(vModel)
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

func (s *InsertRunner) insertRelation(vModel model.Model, vField model.Field) (err *cd.Error) {
	if model.IsSliceField(vField) {
		rErr := s.insertSliceRelation(vModel, vField)
		if rErr != nil {
			err = rErr
			log.Errorf("insertRelation failed, s.insertSliceRelation error:%s", err.Error())
			return
		}
		return
	}

	rErr := s.insertSingleRelation(vModel, vField)
	if rErr != nil {
		err = rErr
		log.Errorf("insertRelation failed, s.insertSingleRelation error:%s", err.Error())
		return
	}
	return
}

func (s *InsertRunner) insertSingleRelation(vModel model.Model, vField model.Field) (err *cd.Error) {
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

	if !model.IsPtrField(vField) {
		rInsertRunner := NewInsertRunner(rModel, s.executor, s.modelProvider, s.modelCodec)
		rModel, rErr = rInsertRunner.Insert()
		if rErr != nil {
			err = rErr
			log.Errorf("insertSingleRelation failed, rInsertRunner.Insert() error:%s", err.Error())
			return
		}
	}

	relationSQL, relationErr := s.hBuilder.BuildInsertRelation(vModel, vField, rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("insertSingleRelation failed, s.hBuilder.BuildInsertRelation error:%s", err.Error())
		return
	}

	var idVal any
	err = s.executor.ExecuteInsert(relationSQL.SQL(), &idVal, relationSQL.Args()...)
	if err != nil {
		log.Errorf("insertSingleRelation failed, s.executor.Execute error:%s", err.Error())
		return
	}

	vField.SetValue(rModel.Interface(model.IsPtrField(vField)))
	return
}

func (s *InsertRunner) insertSliceRelation(vModel model.Model, vField model.Field) (err *cd.Error) {
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
			rInsertRunner := NewInsertRunner(rModel, s.executor, s.modelProvider, s.modelCodec)
			rModel, rErr = rInsertRunner.Insert()
			if rErr != nil {
				err = rErr
				log.Errorf("insertSliceRelation failed, model:%s, filed name:%s, s.insertSingle error:%s", vModel.GetPkgKey(), vField.GetName(), err.Error())
				return
			}
		}

		relationResult, relationErr := s.hBuilder.BuildInsertRelation(vModel, vField, rModel)
		if relationErr != nil {
			err = relationErr
			log.Errorf("insertSliceRelation failed, model:%s, filed name:%s, s.hBuilder.BuildInsertRelation error:%s", vModel.GetPkgKey(), vField.GetName(), err.Error())
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

func (s *InsertRunner) Insert() (ret model.Model, err *cd.Error) {
	err = s.insertHost(s.vModel)
	if err != nil {
		log.Errorf("Insert failed, s.insertSingle error:%s", err.Error())
		return
	}

	for _, field := range s.vModel.GetFields() {
		if model.IsBasicField(field) || !model.IsValidField(field) {
			continue
		}

		err = s.insertRelation(s.vModel, field)
		if err != nil {
			log.Errorf("Insert failed, s.insertRelation error:%s", err.Error())
			return
		}
	}

	ret = s.vModel
	return
}

func (s *impl) Insert(vModel model.Model) (ret model.Model, err *cd.Error) {
	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}

	err = s.executor.BeginTransaction()
	if err != nil {
		return
	}
	defer s.finalTransaction(err)

	insertRunner := NewInsertRunner(vModel, s.executor, s.modelProvider, s.modelCodec)
	ret, err = insertRunner.Insert()
	if err != nil {
		log.Errorf("Insert failed, insertRunner.Insert() error:%s", err.Error())
		return
	}
	return
}
