package orm

import (
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) innerInsert(vModel model.Model) (ret interface{}, err error) {
	builder := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	sqlStr, sqlErr := builder.BuildInsert()
	if sqlErr != nil {
		err = sqlErr
		log.Errorf("innerInsert failed, builder.BuildInsert error:%s", err.Error())
		return
	}

	_, id, idErr := s.executor.Execute(sqlStr)
	if idErr != nil {
		err = idErr
		log.Errorf("innerInsert failed, s.executor.Execute error:%s", err.Error())
		return
	}

	ret = id
	return
}

func (s *impl) insertSingle(vModel model.Model) (err error) {
	autoIncrementFlag := false
	for _, field := range vModel.GetFields() {
		if !field.IsBasic() {
			continue
		}

		fSpec := field.GetSpec()
		fValue := field.GetValue()
		if fValue.IsZero() {
			fValue = s.modelProvider.GetValue(fSpec.GetValueDeclare())
			if !fValue.IsNil() {
				field.SetValue(fValue)
			}
		}
		if fSpec.GetValueDeclare() == model.AutoIncrement {
			autoIncrementFlag = true
		}
	}

	pkVal, pkErr := s.innerInsert(vModel)
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

func (s *impl) insertRelation(vModel model.Model, vField model.Field) (err error) {
	fValue := vField.GetValue()
	if fValue.IsZero() {
		return
	}

	if vField.IsSlice() {
		rValue, rErr := s.insertSliceRelation(vModel, vField)
		if rErr != nil {
			err = rErr
			log.Errorf("insertRelation failed, s.insertSliceRelation error:%s", err.Error())
			return
		}

		vField.SetValue(rValue)
		return
	}

	rValue, rErr := s.insertSingleRelation(vModel, vField)
	if rErr != nil {
		err = rErr
		log.Errorf("insertRelation failed, s.insertSingleRelation error:%s", err.Error())
		return
	}

	vField.SetValue(rValue)
	return
}

func (s *impl) insertSingleRelation(vModel model.Model, vField model.Field) (ret model.Value, err error) {
	fValue := vField.GetValue()
	fType := vField.GetType()
	rModel, rErr := s.modelProvider.GetValueModel(fValue, fType)
	if rErr != nil {
		err = rErr
		log.Errorf("insertSingleRelation failed, s.modelProvider.GetValueModel error:%s", err.Error())
		return
	}

	if !fType.IsPtrType() {
		rErr = s.insertSingle(rModel)
		if rErr != nil {
			err = rErr
			log.Errorf("insertSingleRelation failed, s.insertSingle error:%s", err.Error())
			return
		}

		for _, field := range rModel.GetFields() {
			if field.IsBasic() {
				continue
			}

			err = s.insertRelation(rModel, field)
			if err != nil {
				log.Errorf("insertSingleRelation failed, s.insertRelation error:%s", err.Error())
				return
			}
		}
	}

	builder := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	relationSQL, relationErr := builder.BuildInsertRelation(vField, rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("insertSingleRelation failed, builder.BuildInsertRelation error:%s", err.Error())
		return
	}

	_, _, err = s.executor.Execute(relationSQL)
	if err != nil {
		log.Errorf("insertSingleRelation failed, s.executor.Execute error:%s", err.Error())
		return
	}

	entityVal, entityErr := s.modelProvider.GetEntityValue(rModel.Interface(fType.IsPtrType()))
	if entityErr != nil {
		err = entityErr
		log.Errorf("insertSingleRelation failed, s.modelProvider.GetEntityValue error:%s", err.Error())
		return
	}

	ret = entityVal
	return
}

func (s *impl) insertSliceRelation(vModel model.Model, vField model.Field) (ret model.Value, err error) {
	fValue := vField.GetValue()
	fType := vField.GetType()
	rvValue, _ := fType.Interface(nil)
	fSliceValue, fSliceErr := s.modelProvider.ElemDependValue(fValue)
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
			rErr = s.insertSingle(rModel)
			if rErr != nil {
				err = rErr
				log.Errorf("insertSliceRelation failed, s.insertSingle error:%s", err.Error())
				return
			}

			for _, field := range rModel.GetFields() {
				if field.IsBasic() {
					continue
				}

				err = s.insertRelation(rModel, field)
				if err != nil {
					log.Errorf("insertSliceRelation failed, s.insertRelation error:%s", err.Error())
					return
				}
			}
		}

		builder := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
		relationSQL, relationErr := builder.BuildInsertRelation(vField, rModel)
		if relationErr != nil {
			err = relationErr
			log.Errorf("insertSliceRelation failed, builder.BuildInsertRelation error:%s", err.Error())
			return
		}

		_, _, err = s.executor.Execute(relationSQL)
		if err != nil {
			log.Errorf("insertSliceRelation failed, s.executor.Execute error:%s", err.Error())
			return
		}

		entityVal, entityErr := s.modelProvider.GetEntityValue(rModel.Interface(elemType.IsPtrType()))
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

// Insert insert
func (s *impl) Insert(vModel model.Model) (ret model.Model, err error) {
	err = s.executor.BeginTransaction()
	if err != nil {
		return
	}
	defer s.finalTransaction(err)

	iErr := s.insertSingle(vModel)
	if iErr != nil {
		err = iErr
		log.Errorf("Insert failed, s.insertSingle error:%s", err.Error())
		return
	}

	for _, field := range vModel.GetFields() {
		if field.IsBasic() {
			continue
		}

		err = s.insertRelation(vModel, field)
		if err != nil {
			log.Errorf("Insert failed, s.insertRelation error:%s", err.Error())
			return
		}
	}

	ret = vModel
	return
}
