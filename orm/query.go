package orm

import (
	"fmt"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

type resultItems []interface{}
type resultItemsList []resultItems

func (s *impl) innerQuery(vModel model.Model, filter model.Filter) (ret resultItemsList, err error) {
	builder := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	sqlStr, sqlErr := builder.BuildQuery(filter)
	if sqlErr != nil {
		err = sqlErr
		log.Errorf("build query sql failed, err:%s", err.Error())
		return
	}

	err = s.executor.Query(sqlStr)
	if err != nil {
		return
	}
	defer s.executor.Finish()
	for s.executor.Next() {
		itemValues, itemErr := s.getFieldScanDestPtr(vModel, builder)
		if itemErr != nil {
			err = itemErr
			log.Errorf("getFieldScanDestPtr failed, err:%s", err.Error())
			return
		}

		err = s.executor.GetField(itemValues...)
		if err != nil {
			return
		}

		ret = append(ret, itemValues)
	}

	return
}

func (s *impl) querySingle(vModel model.Model, deepLevel int) (ret model.Model, err error) {
	vFilter, vErr := s.getModelFilter(vModel)
	if vErr != nil {
		err = vErr
		return
	}

	queryValueList, queryErr := s.innerQuery(vModel, vFilter)
	if queryErr != nil {
		err = queryErr
		return
	}

	resultSize := len(queryValueList)
	if resultSize == 0 {
		return
	}
	if resultSize > 1 {
		err = fmt.Errorf("illegal query model, matched to many enties")
		return
	}

	modelVal, modelErr := s.assignSingleModel(vModel, queryValueList[0], deepLevel)
	if modelErr != nil {
		err = modelErr
		return
	}

	ret = modelVal
	return
}

func (s *impl) assignModelField(vModel model.Model, vField model.Field, deepLevel int) (err error) {
	itemVal, itemErr := s.queryRelation(vModel, vField, deepLevel+1)
	if itemErr != nil {
		err = itemErr
		return
	}

	err = vField.SetValue(itemVal)
	if err != nil {
		return
	}
	return
}

func (s *impl) assignBasicField(fType model.Type, vField model.Field, val interface{}) (err error) {
	fVal, fErr := s.modelProvider.DecodeValue(val, fType)
	if fErr != nil {
		err = fErr
		return
	}

	err = vField.SetValue(fVal)
	return
}

func (s *impl) assignSingleModel(modelVal model.Model, queryVal resultItems, deepLevel int) (ret model.Model, err error) {
	offset := 0
	for _, field := range modelVal.GetFields() {
		fType := field.GetType()
		fValue := field.GetValue()
		if !fType.IsBasic() {
			if !fValue.IsNil() {
				err = s.assignModelField(modelVal, field, deepLevel+1)
				if err != nil {
					return
				}
			}

			//offset++
			continue
		}

		if !fValue.IsNil() {
			err = s.assignBasicField(fType, field, queryVal[offset])
			if err != nil {
				return
			}
		}
		offset++
	}

	ret = modelVal
	return
}

func (s *impl) queryRelationSingle(id int, vModel model.Model, deepLevel int) (ret model.Model, err error) {
	relationModel := vModel.Copy()
	relationVal, relationErr := s.modelProvider.GetEntityValue(id)
	if relationErr != nil {
		err = fmt.Errorf("queryRelationSingle failed, GetEntityValue err:%s", relationErr)
		return
	}

	pkField := relationModel.GetPrimaryField()
	pkField.SetValue(relationVal)
	queryVal, queryErr := s.querySingle(relationModel, deepLevel)
	if queryErr != nil {
		err = queryErr
		return
	}

	ret = queryVal
	return
}

func (s *impl) queryRelationSlice(ids []int, vModel model.Model, deepLevel int) (ret []model.Model, err error) {
	sliceVal := []model.Model{}
	for _, item := range ids {
		relationModel := vModel.Copy()
		relationVal, relationErr := s.modelProvider.GetEntityValue(item)
		if relationErr != nil {
			err = fmt.Errorf("GetEntityValue failed, err:%s", relationErr)
			return
		}

		pkField := relationModel.GetPrimaryField()
		pkField.SetValue(relationVal)
		queryVal, queryErr := s.querySingle(relationModel, deepLevel)
		if queryErr != nil {
			err = queryErr
			return
		}

		if queryVal != nil {
			sliceVal = append(sliceVal, queryVal)
		}
	}

	ret = sliceVal
	return
}

func (s *impl) queryRelationIDs(vModel model.Model, rModel model.Model, vField model.Field) (ret []int, err error) {
	builder := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	relationSQL, relationErr := builder.BuildQueryRelation(vField, rModel)
	if relationErr != nil {
		err = relationErr
		return
	}

	var values []int
	func() {
		err = s.executor.Query(relationSQL)
		if err != nil {
			return
		}

		defer s.executor.Finish()
		for s.executor.Next() {
			v := 0
			err = s.executor.GetField(&v)
			if err != nil {
				return
			}
			values = append(values, v)
		}
	}()

	if err != nil {
		return
	}

	ret = values
	return
}

func (s *impl) querySingleRelation(vModel model.Model, vField model.Field, deepLevel int) (ret model.Value, err error) {
	fieldType := vField.GetType()
	fieldModel, fieldErr := s.modelProvider.GetTypeModel(fieldType)
	if fieldErr != nil {
		err = fieldErr
		return
	}

	valuesList, valueErr := s.queryRelationIDs(vModel, fieldModel, vField)
	if valueErr != nil || len(valuesList) == 0 {
		ret = fieldType.Interface()
		return
	}

	fieldValue := fieldType.Interface()
	queryVal, queryErr := s.queryRelationSingle(valuesList[0], fieldModel, deepLevel+1)
	if queryErr != nil {
		err = queryErr
		return
	}

	if queryVal != nil {
		modelVal, modelErr := s.modelProvider.GetEntityValue(queryVal.Interface(true))
		if modelErr != nil {
			err = modelErr
			return
		}

		fieldValue.Set(modelVal.Get())
	}

	ret = fieldValue
	return
}

func (s *impl) querySliceRelation(vModel model.Model, vField model.Field, deepLevel int) (ret model.Value, err error) {
	fieldType := vField.GetType()
	fieldModel, fieldErr := s.modelProvider.GetTypeModel(fieldType)
	if fieldErr != nil {
		err = fieldErr
		return
	}

	valuesList, valueErr := s.queryRelationIDs(vModel, fieldModel, vField)
	if valueErr != nil || len(valuesList) == 0 {
		ret = fieldType.Interface()
		return
	}

	fieldValue := fieldType.Interface()
	queryVal, queryErr := s.queryRelationSlice(valuesList, fieldModel, deepLevel+1)
	if queryErr != nil {
		err = queryErr
		return
	}

	for _, v := range queryVal {
		modelVal, modelErr := s.modelProvider.GetEntityValue(v.Interface(true))
		if modelErr != nil {
			err = modelErr
			return
		}

		fieldValue, fieldErr = s.modelProvider.AppendSliceValue(fieldValue, modelVal)
		if fieldErr != nil {
			err = fieldErr
			return
		}
	}
	ret = fieldValue
	return
}

func (s *impl) queryRelation(vModel model.Model, vField model.Field, deepLevel int) (ret model.Value, err error) {
	fieldType := vField.GetType()
	if deepLevel > maxDeepLevel {
		ret = fieldType.Interface()
		return
	}

	if model.IsStructType(fieldType.GetValue()) {
		ret, err = s.querySingleRelation(vModel, vField, deepLevel)
		return
	}

	if model.IsSliceType(fieldType.GetValue()) {
		ret, err = s.querySliceRelation(vModel, vField, deepLevel)
		return
	}

	return
}

func (s *impl) Query(vModel model.Model) (ret model.Model, err error) {
	queryVal, queryErr := s.querySingle(vModel, 0)
	if queryErr != nil {
		err = queryErr
		return
	}

	if queryVal != nil {
		ret = queryVal
		return
	}

	err = fmt.Errorf("not exist model,name:%s,pkgPath:%s", vModel.GetName(), vModel.GetPkgPath())
	return
}
