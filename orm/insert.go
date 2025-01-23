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

func (s *InsertRunner) insertHost(vModel model.Model) (err *cd.Result) {
	autoIncrementFlag := false
	for _, field := range vModel.GetFields() {
		if !field.IsBasic() {
			continue
		}

		fSpec := field.GetSpec()
		fValue := field.GetValue()
		if !fValue.IsValid() {
			fValue = s.modelProvider.GetNewValue(fSpec.GetValueDeclare())
			field.SetValue(fValue)
		}
		if fSpec.GetValueDeclare() == model.AutoIncrement {
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
		pkField := vModel.GetPrimaryField()
		tVal, tErr := s.modelProvider.DecodeValue(pkVal, pkField.GetType())
		if tErr != nil {
			err = tErr
			log.Errorf("insertHost failed, s.modelProvider.DecodeValue error:%s", err.Error())
			return
		}

		pkField.SetValue(tVal)
	}
	return
}

func (s *InsertRunner) innerHost(vModel model.Model) (ret model.RawVal, err *cd.Result) {
	insertResult, insertErr := s.hBuilder.BuildInsert(vModel)
	if insertErr != nil {
		err = insertErr
		log.Errorf("innerHost failed, builder.BuildInsert error:%s", err.Error())
		return
	}

	_, id, idErr := s.executor.Execute(insertResult.SQL(), insertResult.Args()...)
	if idErr != nil {
		err = idErr
		log.Errorf("innerHost failed, s.executor.Execute error:%s", err.Error())
		return
	}

	ret = model.NewRawVal(id)
	return
}

func (s *InsertRunner) insertRelation(vModel model.Model, vField model.Field, rModel model.Model) (err *cd.Result) {
	fValue := vField.GetValue()
	if fValue.IsZero() {
		return
	}

	if vField.IsSlice() {
		rValue, rErr := s.insertSliceRelation(vModel, vField, rModel)
		if rErr != nil {
			err = rErr
			log.Errorf("insertRelation failed, s.insertSliceRelation error:%s", err.Error())
			return
		}

		if rValue != nil && !rValue.IsZero() {
			vField.SetValue(rValue)
		}
		return
	}

	rValue, rErr := s.insertSingleRelation(vModel, vField, rModel)
	if rErr != nil {
		err = rErr
		log.Errorf("insertRelation failed, s.insertSingleRelation error:%s", err.Error())
		return
	}

	if rValue != nil && !rValue.IsZero() {
		vField.SetValue(rValue)
	}
	return
}

func (s *InsertRunner) insertSingleRelation(vModel model.Model, vField model.Field, rModel model.Model) (ret model.Value, err *cd.Result) {
	fType := vField.GetType()
	fValue := vField.GetValue()
	var rErr *cd.Result
	rModel, rErr = s.modelProvider.SetModelValue(rModel, fValue)
	if rErr != nil {
		err = rErr
		log.Errorf("insertSingleRelation failed, s.modelProvider.SetModelValue error:%s", err.Error())
		return
	}

	if !vField.IsPtrType() {
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

	_, _, err = s.executor.Execute(relationSQL.SQL(), relationSQL.Args()...)
	if err != nil {
		log.Errorf("insertSingleRelation failed, s.executor.Execute error:%s", err.Error())
		return
	}

	entityVal, entityErr := s.modelProvider.GetEntityValue(rModel.Interface(fType.IsPtrType(), model.LiteView))
	if entityErr != nil {
		err = entityErr
		log.Errorf("insertSingleRelation failed, s.modelProvider.GetEntityValue error:%s", err.Error())
		return
	}

	ret = entityVal
	return
}

func (s *InsertRunner) insertSliceRelation(vModel model.Model, vField model.Field, rModel model.Model) (ret model.Value, err *cd.Result) {
	fValue := vField.GetValue()
	fType := vField.GetType()
	rvValue, _ := fType.Interface(nil)
	fSliceValue, fSliceErr := s.modelProvider.ElemDependValue(fValue.Interface())
	if fSliceErr != nil {
		err = fSliceErr
		log.Errorf("insertSliceRelation failed, s.modelProvider.ElemDependValue error:%s", err.Error())
		return
	}

	var rErr *cd.Result
	for _, fVal := range fSliceValue {
		rModel, rErr = s.modelProvider.SetModelValue(rModel.Copy(true), fVal)
		if rErr != nil {
			err = rErr
			log.Errorf("insertSliceRelation failed, s.modelProvider.GetValueModel error:%s", err.Error())
			return
		}

		elemType := fType.Elem()
		if !elemType.IsPtrType() {
			rInsertRunner := NewInsertRunner(rModel, s.executor, s.modelProvider, s.modelCodec)
			rModel, rErr = rInsertRunner.Insert()
			if rErr != nil {
				err = rErr
				log.Errorf("insertSliceRelation failed, s.insertSingle error:%s", err.Error())
				return
			}
		}

		relationResult, relationErr := s.hBuilder.BuildInsertRelation(vModel, vField, rModel)
		if relationErr != nil {
			err = relationErr
			log.Errorf("insertSliceRelation failed, s.hBuilder.BuildInsertRelation error:%s", err.Error())
			return
		}

		_, _, err = s.executor.Execute(relationResult.SQL(), relationResult.Args()...)
		if err != nil {
			log.Errorf("insertSliceRelation failed, s.executor.Execute error:%s", err.Error())
			return
		}

		entityVal, entityErr := s.modelProvider.GetEntityValue(rModel.Interface(elemType.IsPtrType(), model.LiteView))
		if entityErr != nil {
			err = entityErr
			log.Errorf("insertSliceRelation failed, s.modelProvider.GetEntityValue error:%s", err.Error())
			return
		}

		rvValue, err = s.modelProvider.AppendSliceValue(rvValue, entityVal)
		if err != nil {
			log.Errorf("insertSliceRelation failed, s.modelProvider.AppendSliceValue error:%s", err.Error())
			return
		}
	}

	ret = rvValue
	return
}

func (s *InsertRunner) Insert() (ret model.Model, err *cd.Result) {
	err = s.insertHost(s.vModel)
	if err != nil {
		log.Errorf("Insert failed, s.insertSingle error:%s", err.Error())
		return
	}

	for _, field := range s.vModel.GetFields() {
		if field.IsBasic() {
			continue
		}

		rModel, rErr := s.modelProvider.GetTypeModel(field.GetType())
		if rErr != nil {
			err = rErr
			log.Errorf("Insert failed, s.modelProvider.GetTypeModel error:%s", err.Error())
			return
		}

		err = s.insertRelation(s.vModel, field, rModel)
		if err != nil {
			log.Errorf("Insert failed, s.insertRelation error:%s", err.Error())
			return
		}
	}

	ret = s.vModel
	return
}

func (s *impl) Insert(vModel model.Model) (ret model.Model, err *cd.Result) {
	if vModel == nil {
		err = cd.NewResult(cd.IllegalParam, "illegal model value")
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
