package orm

import (
	"fmt"

	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

type resultItems []any
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
		itemValues, itemErr := s.getModelFieldsScanDestPtr(vModel, builder)
		if itemErr != nil {
			err = itemErr
			log.Errorf("innerQuery failed, s.getModelFieldsScanDestPtr error:%s", err.Error())
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

func (s *impl) innerAssign(vModel model.Model, queryVal resultItems, deepLevel int) (ret model.Model, err error) {
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
				log.Errorf("innerAssign failed, s.assignModelField error:%v", err.Error())
				return
			}

			//offset++
			continue
		}

		err = s.assignBasicField(field, fType, queryVal[offset])
		if err != nil {
			log.Errorf("innerAssign failed, s.assignBasicField error:%v", err.Error())
			return
		}
		offset++
	}

	ret = vModel
	return
}

func (s *impl) querySingle(vModel model.Model, vFilter model.Filter, deepLevel int) (ret model.Model, err error) {
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

	modelVal, modelErr := s.innerAssign(vModel, valueList[0], deepLevel)
	if modelErr != nil {
		err = modelErr
		log.Errorf("querySingle failed, s.innerAssign error:%v", err.Error())
		return
	}

	ret = modelVal
	return
}

func (s *impl) assignModelField(vField model.Field, vModel model.Model, deepLevel int) (err error) {
	vVal, vErr := s.queryRelation(vModel, vField, deepLevel)
	if vErr != nil {
		err = vErr
		log.Errorf("assignModelField failed, s.queryRelation error:%v", err.Error())
		return
	}

	vField.SetValue(vVal)
	return
}

func (s *impl) assignBasicField(vField model.Field, fType model.Type, val interface{}) (err error) {
	fVal, fErr := s.modelProvider.DecodeValue(val, fType)
	if fErr != nil {
		err = fErr
		log.Errorf("assignBasicField failed, s.modelProvider.DecodeValue error:%v", err.Error())
		return
	}

	vField.SetValue(fVal)
	return
}

func (s *impl) innerQueryRelationModel(id any, vModel model.Model, deepLevel int) (ret model.Model, err error) {
	vFilter, vErr := s.getModelFilter(vModel)
	if vErr != nil {
		err = vErr
		log.Errorf("innerQueryRelationModel failed, getModelFilter error:%v", err.Error())
		return
	}

	pkField := vModel.GetPrimaryField()
	fVal, fErr := s.modelProvider.DecodeValue(id, pkField.GetType())
	if fErr != nil {
		err = fErr
		log.Errorf("innerQueryRelationModel failed, s.modelProvider.DecodeValue error:%v", err.Error())
		return
	}

	err = vFilter.Equal(pkField.GetName(), fVal.Interface())
	if err != nil {
		log.Errorf("innerQueryRelationModel failed, vFilter.Equal error:%v", err.Error())
		return
	}

	queryVal, queryErr := s.querySingle(vModel, vFilter, deepLevel+1)
	if queryErr != nil {
		err = queryErr
		log.Errorf("innerQueryRelationModel failed, s.querySingle error:%v", err.Error())
		return
	}

	ret = queryVal
	return
}

func (s *impl) innerQueryRelationSliceModel(ids []any, vModel model.Model, deepLevel int) (ret []model.Model, err error) {
	sliceVal := []model.Model{}
	for _, id := range ids {
		vFilter, vErr := s.getModelFilter(vModel)
		if vErr != nil {
			err = vErr
			log.Errorf("innerQueryRelationSliceModel failed, getModelFilter error:%v", err.Error())
			return
		}

		pkField := vModel.GetPrimaryField()
		fVal, fErr := s.modelProvider.DecodeValue(id, pkField.GetType())
		if fErr != nil {
			err = fErr
			log.Errorf("innerQueryRelationSliceModel failed, s.modelProvider.DecodeValue error:%v", err.Error())
			return
		}

		err = vFilter.Equal(pkField.GetName(), fVal.Interface())
		if err != nil {
			log.Errorf("innerQueryRelationSliceModel failed, vFilter.Equal error:%v", err.Error())
			return
		}
		queryVal, queryErr := s.querySingle(vModel.Copy(false), vFilter, deepLevel+1)
		if queryErr != nil {
			err = queryErr
			log.Errorf("innerQueryRelationSliceModel failed, s.querySingle error:%v", err.Error())
			return
		}

		if queryVal != nil {
			sliceVal = append(sliceVal, queryVal)
		}
	}

	ret = sliceVal
	return
}

func (s *impl) innerQueryRelationKeys(vModel model.Model, rModel model.Model, vField model.Field) (ret resultItems, err error) {
	builder := builder.NewBuilder(vModel, s.modelProvider, s.specialPrefix)
	relationSQL, relationErr := builder.BuildQueryRelation(vField, rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("innerQueryRelationKeys failed, builder.BuildQueryRelation error:%v", err.Error())
		return
	}

	values := resultItems{}
	func() {
		err = s.executor.Query(relationSQL)
		if err != nil {
			log.Errorf("innerQueryRelationKeys failed, s.executor.Query error:%v", err.Error())
			return
		}

		defer s.executor.Finish()
		for s.executor.Next() {
			itemValue, itemErr := s.getModelPKFieldScanDestPtr(vModel, builder)
			if itemErr != nil {
				err = itemErr
				log.Errorf("innerQueryRelationKeys failed, s.getModelPKFieldScanDestPtr error:%v", err.Error())
				return
			}

			err = s.executor.GetField(itemValue)
			if err != nil {
				log.Errorf("innerQueryRelationKeys failed, s.executor.GetField error:%v", err.Error())
				return
			}
			values = append(values, itemValue)
		}
	}()

	if err != nil {
		return
	}

	ret = values
	return
}

func (s *impl) querySingleRelation(vModel model.Model, vField model.Field, deepLevel int) (ret model.Value, err error) {
	rModel, rErr := s.modelProvider.GetTypeModel(vField.GetType())
	if rErr != nil {
		err = rErr
		log.Errorf("querySingleRelation failed, s.modelProvider.GetTypeModel err:%v", err.Error())
		return
	}

	vType := vField.GetType()
	vValue, _ := vType.Interface(nil)
	valueList, valueErr := s.innerQueryRelationKeys(vModel, rModel, vField)
	if valueErr != nil {
		err = valueErr
		log.Errorf("querySingleRelation failed, s.innerQueryRelationKeys error:%v", err.Error())
		return
	}
	valueSize := len(valueList)
	if valueSize == 0 {
		if vType.IsPtrType() {
			ret = vValue
			return
		}

		err = fmt.Errorf("mismatch relation field:%s", vField.GetName())
		return
	}

	rModel, rErr = s.innerQueryRelationModel(valueList[0], rModel, deepLevel)
	if rErr != nil {
		err = rErr
		log.Errorf("querySingleRelation failed, s.innerQueryRelationModel error:%v", err.Error())
		return
	}

	modelVal, modelErr := s.modelProvider.GetEntityValue(rModel.Interface(vType.IsPtrType()))
	if modelErr != nil {
		err = modelErr
		log.Errorf("querySingleRelation failed, s.modelProvider.GetEntityValue err:%v", err.Error())
		return
	}

	vValue.Set(modelVal.Get())
	ret = vValue
	return
}

func (s *impl) querySliceRelation(vModel model.Model, vField model.Field, deepLevel int) (ret model.Value, err error) {
	rModel, rErr := s.modelProvider.GetTypeModel(vField.GetType())
	if rErr != nil {
		err = rErr
		log.Errorf("querySliceRelation failed, s.modelProvider.GetTypeModel err:%sv", err.Error())
		return
	}

	vType := vField.GetType()
	vValue, _ := vType.Interface(nil)
	valueList, valueErr := s.innerQueryRelationKeys(vModel, rModel, vField)
	if valueErr != nil {
		err = valueErr
		log.Errorf("querySliceRelation failed, s.innerQueryRelationKeys err:%sv", err.Error())
		return
	}
	valueSize := len(valueList)
	if valueSize == 0 {
		ret = vValue
		return
	}

	rModelList, rModelErr := s.innerQueryRelationSliceModel(valueList, rModel, deepLevel)
	if rModelErr != nil {
		err = rModelErr
		log.Errorf("querySliceRelation failed, s.innerQueryRelationSliceModel err:%sv", err.Error())
		return
	}

	elemType := vType.Elem()
	for _, sv := range rModelList {
		modelVal, modelErr := s.modelProvider.GetEntityValue(sv.Interface(elemType.IsPtrType()))
		if modelErr != nil {
			err = modelErr
			log.Errorf("querySliceRelation failed, s.modelProvider.GetEntityValue err:%sv", err.Error())
			return
		}

		vValue, rErr = s.modelProvider.AppendSliceValue(vValue, modelVal)
		if rErr != nil {
			err = rErr
			log.Errorf("querySliceRelation failed, s.modelProvider.AppendSliceValue err:%sv", err.Error())
			return
		}
	}
	ret = vValue
	return
}

func (s *impl) queryRelation(vModel model.Model, vField model.Field, deepLevel int) (ret model.Value, err error) {
	if deepLevel > maxDeepLevel {
		fieldType := vField.GetType()
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
	if vModel == nil {
		err = fmt.Errorf("illegal model value")
		return
	}

	vFilter, vErr := s.getModelFilter(vModel)
	if vErr != nil {
		err = vErr
		log.Errorf("Query failed, s.getModelFilter error:%v", err.Error())
		return
	}

	qModel, qErr := s.querySingle(vModel, vFilter, 0)
	if qErr != nil {
		err = qErr
		log.Errorf("Query failed, s.querySingle err:%v", err.Error())
		return
	}

	ret = qModel
	return
}
