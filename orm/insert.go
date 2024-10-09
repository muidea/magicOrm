package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/database/context"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) innerInsert(builder builder.Builder) (ret model.RawVal, err *cd.Result) {
	insertResult, insertErr := builder.BuildInsert()
	if insertErr != nil {
		err = insertErr
		log.Errorf("innerInsert failed, builder.BuildInsert error:%s", err.Error())
		return
	}

	_, id, idErr := s.executor.Execute(insertResult.SQL(), insertResult.Args()...)
	if idErr != nil {
		err = idErr
		log.Errorf("innerInsert failed, s.executor.Execute error:%s", err.Error())
		return
	}

	ret = model.NewRawVal(id)
	return
}

func (s *impl) insertSingle(builder builder.Builder, vModel model.Model) (err *cd.Result) {
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

	pkVal, pkErr := s.innerInsert(builder)
	if pkErr != nil {
		err = pkErr
		log.Errorf("insertSingle failed, s.innerInsert error:%s", err.Error())
		return
	}

	if pkVal != nil && autoIncrementFlag {
		pkField := vModel.GetPrimaryField()
		tVal, tErr := s.modelProvider.DecodeValue(pkVal, pkField.GetType())
		if tErr != nil {
			err = tErr
			log.Errorf("insertSingle failed, s.modelProvider.DecodeValue error:%s", err.Error())
			return
		}

		pkField.SetValue(tVal)
	}
	return
}

func (s *impl) insertRelation(builder builder.Builder, vField model.Field) (err *cd.Result) {
	fValue := vField.GetValue()
	if fValue.IsZero() {
		return
	}

	if vField.IsSlice() {
		rValue, rErr := s.insertSliceRelation(builder, vField)
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

	rValue, rErr := s.insertSingleRelation(builder, vField)
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

func (s *impl) insertSingleRelation(hBuilder builder.Builder, vField model.Field) (ret model.Value, err *cd.Result) {
	fValue := vField.GetValue()
	fType := vField.GetType()
	rModel, rErr := s.modelProvider.GetValueModel(fValue, fType)
	if rErr != nil {
		err = rErr
		log.Errorf("insertSingleRelation failed, s.modelProvider.GetValueModel error:%s", err.Error())
		return
	}

	if !fType.IsPtrType() {
		rContext := context.New(rModel, s.modelProvider, s.specialPrefix)
		rBuilder := builder.NewBuilder(rModel, rContext)
		rErr = s.insertSingle(rBuilder, rModel)
		if rErr != nil {
			err = rErr
			log.Errorf("insertSingleRelation failed, s.insertSingle error:%s", err.Error())
			return
		}

		for _, field := range rModel.GetFields() {
			if field.IsBasic() {
				continue
			}

			err = s.insertRelation(rBuilder, field)
			if err != nil {
				log.Errorf("insertSingleRelation failed, s.insertRelation error:%s", err.Error())
				return
			}
		}
	}

	relationSQL, relationErr := hBuilder.BuildInsertRelation(vField, rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("insertSingleRelation failed, builderVal.BuildInsertRelation error:%s", err.Error())
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

func (s *impl) insertSliceRelation(hBuilder builder.Builder, vField model.Field) (ret model.Value, err *cd.Result) {
	fValue := vField.GetValue()
	fType := vField.GetType()
	rvValue, _ := fType.Interface(nil)
	fSliceValue, fSliceErr := s.modelProvider.ElemDependValue(fValue.Interface())
	if fSliceErr != nil {
		err = fSliceErr
		log.Errorf("insertSliceRelation failed, s.modelProvider.ElemDependValue error:%s", err.Error())
		return
	}

	elemType := fType.Elem()
	for _, fVal := range fSliceValue {
		rModel, rErr := s.modelProvider.GetValueModel(fVal, elemType)
		if rErr != nil {
			err = rErr
			log.Errorf("insertSliceRelation failed, s.modelProvider.GetValueModel error:%s", err.Error())
			return
		}

		if !elemType.IsPtrType() {
			rContext := context.New(rModel, s.modelProvider, s.specialPrefix)
			rBuilder := builder.NewBuilder(rModel, rContext)
			rErr = s.insertSingle(rBuilder, rModel)
			if rErr != nil {
				err = rErr
				log.Errorf("insertSliceRelation failed, s.insertSingle error:%s", err.Error())
				return
			}

			for _, field := range rModel.GetFields() {
				if field.IsBasic() {
					continue
				}

				err = s.insertRelation(rBuilder, field)
				if err != nil {
					log.Errorf("insertSliceRelation failed, s.insertRelation error:%s", err.Error())
					return
				}
			}
		}

		relationResult, relationErr := hBuilder.BuildInsertRelation(vField, rModel)
		if relationErr != nil {
			err = relationErr
			log.Errorf("insertSliceRelation failed, builderVal.BuildInsertRelation error:%s", err.Error())
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

func (s *impl) insertModel(vModel model.Model) (ret model.Model, err *cd.Result) {
	hContext := context.New(vModel, s.modelProvider, s.specialPrefix)
	hBuilder := builder.NewBuilder(vModel, hContext)
	err = s.insertSingle(hBuilder, vModel)
	if err != nil {
		log.Errorf("Insert failed, s.insertSingle error:%s", err.Error())
		return
	}

	for _, field := range vModel.GetFields() {
		if field.IsBasic() {
			continue
		}

		err = s.insertRelation(hBuilder, field)
		if err != nil {
			log.Errorf("Insert failed, s.insertRelation error:%s", err.Error())
			return
		}
	}

	ret = vModel
	return
}

func (s *impl) Insert(vModel model.Model) (ret model.Model, err *cd.Result) {
	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}

	err = s.executor.BeginTransaction()
	if err != nil {
		return
	}
	defer s.finalTransaction(err)

	ret, err = s.insertModel(vModel)
	if err != nil {
		log.Errorf("Insert failed, s.insertModel error:%s", err.Error())
		return
	}
	return
}
