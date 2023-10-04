package orm

import (
	"fmt"

	"github.com/muidea/magicCommon/foundation/log"

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
		log.Errorf("innerQuery failed, builder.BuildQuery error:%s", err.Error())
		return
	}

	err = s.executor.Query(sqlStr)
	if err != nil {
		log.Errorf("innerQuery failed, s.executor.Query error:%s", err.Error())
		return
	}
	defer s.executor.Finish()
	for s.executor.Next() {
		itemValues, itemErr := s.getFieldScanDestPtr(vModel, builder)
		if itemErr != nil {
			err = itemErr
			log.Errorf("innerQuery failed, s.getFieldScanDestPtr error:%s", err.Error())
			return
		}

		err = s.executor.GetField(itemValues...)
		if err != nil {
			log.Errorf("innerQuery failed, s.executor.GetField error:%s", err.Error())
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
		log.Errorf("querySingle failed, getModelFilter error:%v", err.Error())
		return
	}

	valueList, queryErr := s.innerQuery(vModel, vFilter)
	if queryErr != nil {
		err = queryErr
		log.Errorf("querySingle failed, innerQuery error:%v", err.Error())
		return
	}

	resultSize := len(valueList)
	if resultSize != 1 {
		err = fmt.Errorf("matched %d values", resultSize)
		log.Errorf("querySingle failed, error:%v", err.Error())
		return
	}

	modelVal, modelErr := s.assignSingleModel(vModel, valueList[0], deepLevel)
	if modelErr != nil {
		err = modelErr
		log.Errorf("querySingle failed, s.assignSingleModel error:%v", err.Error())
		return
	}

	ret = modelVal
	return
}

func (s *impl) assignModelField(vField model.Field, vModel model.Model, deepLevel int) (err error) {
	itemVal, itemErr := s.queryRelation(vModel, vField, deepLevel)
	if itemErr != nil {
		err = itemErr
		log.Errorf("assignModelField failed, s.queryRelation error:%v", err.Error())
		return
	}

	err = vField.SetValue(itemVal)
	if err != nil {
		log.Errorf("assignModelField failed, vField.SetValue error:%v", err.Error())
		return
	}
	return
}

func (s *impl) assignBasicField(vField model.Field, fType model.Type, val interface{}) (err error) {
	fVal, fErr := s.modelProvider.DecodeValue(val, fType)
	if fErr != nil {
		err = fErr
		log.Errorf("assignBasicField failed, s.modelProvider.DecodeValue error:%v", err.Error())
		return
	}

	err = vField.SetValue(fVal)
	if err != nil {
		log.Errorf("assignBasicField failed, vField.SetValue error:%v", err.Error())
	}
	return
}

func (s *impl) assignSingleModel(vModel model.Model, queryVal resultItems, deepLevel int) (ret model.Model, err error) {
	offset := 0
	for _, field := range vModel.GetFields() {
		fType := field.GetType()
		fValue := field.GetValue()
		if fValue.IsNil() {
			continue
		}

		if !fType.IsBasic() {
			err = s.assignModelField(field, vModel, deepLevel)
			if err != nil {
				log.Errorf("assignSingleModel failed, s.assignModelField error:%v", err.Error())
				return
			}

			//offset++
			continue
		}

		err = s.assignBasicField(field, fType, queryVal[offset])
		if err != nil {
			log.Errorf("assignSingleModel failed, s.assignBasicField error:%v", err.Error())
			return
		}
		offset++
	}

	ret = vModel
	return
}

func (s *impl) queryRelationSingle(id int, vModel model.Model, deepLevel int) (ret model.Model, err error) {
	relationModel := vModel.Copy()
	relationVal, relationErr := s.modelProvider.GetEntityValue(id)
	if relationErr != nil {
		err = relationErr
		log.Errorf("queryRelationSingle failed, s.modelProvider.GetEntityValue error:%v", err.Error())
		return
	}

	pkField := relationModel.GetPrimaryField()
	pkField.SetValue(relationVal)
	queryVal, queryErr := s.querySingle(relationModel, deepLevel+1)
	if queryErr != nil {
		err = queryErr
		log.Errorf("queryRelationSingle failed, s.querySingle error:%v", err.Error())
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
			err = relationErr
			log.Errorf("queryRelationSlice failed, s.modelProvider.GetEntityValue error:%v", err.Error())
			return
		}

		pkField := relationModel.GetPrimaryField()
		pkField.SetValue(relationVal)
		queryVal, queryErr := s.querySingle(relationModel, deepLevel+1)
		if queryErr != nil {
			err = queryErr
			log.Errorf("queryRelationSlice failed, s.querySingle error:%v", err.Error())
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
		log.Errorf("queryRelationIDs failed, builder.BuildQueryRelation error:%v", err.Error())
		return
	}

	var values []int
	func() {
		err = s.executor.Query(relationSQL)
		if err != nil {
			log.Errorf("queryRelationIDs failed, s.executor.Query error:%v", err.Error())
			return
		}

		defer s.executor.Finish()
		for s.executor.Next() {
			v := 0
			err = s.executor.GetField(&v)
			if err != nil {
				log.Errorf("queryRelationIDs failed, s.executor.GetField error:%v", err.Error())
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
	fieldModel, fieldErr := s.modelProvider.GetTypeModel(vField.GetType())
	if fieldErr != nil {
		err = fieldErr
		log.Errorf("querySingleRelation failed, s.modelProvider.GetTypeModel err:%v", err.Error())
		return
	}

	fieldType := vField.GetType()
	fieldValue, _ := fieldType.Interface(nil)
	valueList, valueErr := s.queryRelationIDs(vModel, fieldModel, vField)
	if valueErr != nil {
		err = valueErr
		log.Errorf("querySingleRelation failed, s.queryRelationIDs error:%v", err.Error())
		return
	}
	valueSize := len(valueList)
	if valueSize == 0 {
		if fieldType.IsPtrType() {
			ret = fieldValue
			return
		}

		err = fmt.Errorf("query relation field:%s failed", vField.GetName())
		return
	}

	if valueSize > 1 {
		err = fmt.Errorf("query relation field:%s match %d values", vField.GetName(), valueSize)
		log.Errorf("querySingleRelation failed, s.queryRelationIDs error:%v", err.Error())
		return
	}

	queryVal, queryErr := s.queryRelationSingle(valueList[0], fieldModel, deepLevel)
	if queryErr != nil {
		err = queryErr
		log.Errorf("querySingleRelation failed, s.queryRelationSingle err:%v", err.Error())
		return
	}

	modelVal, modelErr := s.modelProvider.GetEntityValue(queryVal.Interface(fieldType.IsPtrType()))
	if modelErr != nil {
		err = modelErr
		log.Errorf("querySingleRelation failed, s.modelProvider.GetEntityValue err:%v", err.Error())
		return
	}

	err = fieldValue.Set(modelVal.Get())
	if err != nil {
		log.Errorf("querySingleRelation failed, fieldValue.Set err:%v", err.Error())
		return
	}

	ret = fieldValue
	return
}

func (s *impl) querySliceRelation(vModel model.Model, vField model.Field, deepLevel int) (ret model.Value, err error) {
	fieldModel, fieldErr := s.modelProvider.GetTypeModel(vField.GetType())
	if fieldErr != nil {
		err = fieldErr
		log.Errorf("querySliceRelation failed, s.modelProvider.GetTypeModel err:%v", err.Error())
		return
	}

	fieldType := vField.GetType()
	fieldValue, _ := fieldType.Interface(nil)
	valuesList, valueErr := s.queryRelationIDs(vModel, fieldModel, vField)
	if valueErr != nil {
		err = valueErr
		log.Errorf("querySliceRelation failed, s.queryRelationIDs err:%v", err.Error())
		return
	}
	if len(valuesList) == 0 {
		ret = fieldValue
		return
	}

	queryVal, queryErr := s.queryRelationSlice(valuesList, fieldModel, deepLevel+1)
	if queryErr != nil {
		err = queryErr
		log.Errorf("querySliceRelation failed, s.queryRelationSlice err:%v", err.Error())
		return
	}

	elemType := fieldType.Elem()
	for _, v := range queryVal {
		modelVal, modelErr := s.modelProvider.GetEntityValue(v.Interface(elemType.IsPtrType()))
		if modelErr != nil {
			err = modelErr
			log.Errorf("querySliceRelation failed, s.modelProvider.GetEntityValue err:%v", err.Error())
			return
		}

		fieldValue, fieldErr = s.modelProvider.AppendSliceValue(fieldValue, modelVal)
		if fieldErr != nil {
			err = fieldErr
			log.Errorf("querySliceRelation failed, s.modelProvider.AppendSliceValue err:%v", err.Error())
			return
		}
	}
	ret = fieldValue
	return
}

func (s *impl) queryRelation(vModel model.Model, vField model.Field, deepLevel int) (ret model.Value, err error) {
	fieldType := vField.GetType()
	if deepLevel > maxDeepLevel {
		ret, _ = fieldType.Interface(nil)
		return
	}

	if vField.IsSlice() {
		ret, err = s.querySliceRelation(vModel, vField, deepLevel)
		if err != nil {
			log.Errorf("queryRelation failed, s.querySliceRelation err:%v", err.Error())
		}
		return
	}

	ret, err = s.querySingleRelation(vModel, vField, deepLevel)
	if err != nil {
		log.Errorf("queryRelation failed, s.querySingleRelation err:%v", err.Error())
	}
	return
}

func (s *impl) Query(vModel model.Model) (ret model.Model, err error) {
	queryVal, queryErr := s.querySingle(vModel, 0)
	if queryErr != nil {
		err = queryErr
		log.Errorf("Query failed, s.querySingle err:%v", err.Error())
		return
	}

	ret = queryVal
	return
}
