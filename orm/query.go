package orm

import (
	"fmt"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

type resultItems []interface{}
type resultItemsList []resultItems

func (s *impl) innerQuery(elemModel model.Model, filter model.Filter) (ret resultItemsList, err error) {
	builder := builder.NewBuilder(elemModel, s.modelProvider)
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
		itemValues, itemErr := s.getInitializeValue(elemModel, builder)
		if itemErr != nil {
			err = itemErr
			log.Errorf("getInitializeValue failed, err:%s", err.Error())
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

func (s *impl) querySingleModel(vModel model.Model, filter model.Filter, deepLevel int) (ret model.Model, err error) {
	queryValueList, queryErr := s.innerQuery(vModel, filter)
	if queryErr != nil {
		err = queryErr
		return
	}

	resultSize := len(queryValueList)
	if resultSize == 0 {
		return
	}
	if resultSize > 1 {
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

func (s *impl) assignField(vModel model.Model, vField model.Field, queryVal resultItems, deepLevel int) (ret resultItems, err error) {
	fType := vField.GetType()
	fValue := vField.GetValue()
	if !fType.IsBasic() {
		if !fValue.IsNil() {
			itemVal, itemErr := s.queryRelation(vModel, vField, deepLevel+1)
			if itemErr != nil {
				err = itemErr
				return
			}

			err = vField.SetValue(itemVal)
			if err != nil {
				return
			}
		}

		ret = queryVal
		return
	}

	if !fValue.IsNil() {
		fVal, fErr := s.modelProvider.DecodeValue(s.stripSlashes(fType, queryVal[0]), fType)
		if fErr != nil {
			err = fErr
			return
		}
		err = vField.SetValue(fVal)
		if err != nil {
			return
		}

		ret = queryVal[1:]
	}

	ret = queryVal
	return
}

func (s *impl) assignSingleModel(modelVal model.Model, queryVal resultItems, deepLevel int) (ret model.Model, err error) {
	offset := 0
	for _, field := range modelVal.GetFields() {
		fType := field.GetType()
		fValue := field.GetValue()
		if !fType.IsBasic() {
			if !fValue.IsNil() {
				itemVal, itemErr := s.queryRelation(modelVal, field, deepLevel+1)
				if itemErr != nil {
					err = itemErr
					return
				}

				err = field.SetValue(itemVal)
				if err != nil {
					return
				}
			}

			//offset++
			continue
		}

		if !fValue.IsNil() {
			fVal, fErr := s.modelProvider.DecodeValue(s.stripSlashes(fType, queryVal[offset]), fType)
			if fErr != nil {
				err = fErr
				return
			}
			err = field.SetValue(fVal)
			if err != nil {
				return
			}
		}

		offset++
	}

	ret = modelVal
	return
}

func (s *impl) queryRelationSingleModel(id int, vModel model.Model, deepLevel int) (ret model.Model, err error) {
	relationModel := vModel.Copy()
	relationVal, relationErr := s.modelProvider.GetEntityValue(id)
	if relationErr != nil {
		err = fmt.Errorf("queryRelationSingleModel failed, GetEntityValue err:%s", relationErr)
		return
	}

	pkField := relationModel.GetPrimaryField()
	pkField.SetValue(relationVal)
	//relationFilter, relationErr := s.getFieldFilter(pkField)
	relationFilter, relationErr := s.getModelFilter(relationModel)
	if relationErr != nil {
		err = fmt.Errorf("queryRelationSingleModel failed, getModelFilter err:%s", relationErr)
		return
	}

	queryVal, queryErr := s.querySingleModel(relationModel, relationFilter, deepLevel)
	if queryErr != nil {
		err = queryErr
		return
	}

	ret = queryVal
	return
}

func (s *impl) queryRelationSliceModel(ids []int, vModel model.Model, deepLevel int) (ret []model.Model, err error) {
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
		//relationFilter, relationErr := s.getFieldFilter(pkField)
		relationFilter, relationErr := s.getModelFilter(relationModel)
		if relationErr != nil {
			err = fmt.Errorf("GetEntityValue failed, err:%s", relationErr)
			return
		}

		queryVal, queryErr := s.querySingleModel(relationModel, relationFilter, deepLevel)
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
	builder := builder.NewBuilder(vModel, s.modelProvider)
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
	queryVal, queryErr := s.queryRelationSingleModel(valuesList[0], fieldModel, deepLevel+1)
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
	queryVal, queryErr := s.queryRelationSliceModel(valuesList, fieldModel, deepLevel+1)
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

	if util.IsStructType(fieldType.GetValue()) {
		ret, err = s.querySingleRelation(vModel, vField, deepLevel)
		return
	}

	if util.IsSliceType(fieldType.GetValue()) {
		ret, err = s.querySliceRelation(vModel, vField, deepLevel)
		return
	}

	return
}

func (s *impl) Query(entityModel model.Model) (ret model.Model, err error) {
	entityFilter, entityErr := s.getModelFilter(entityModel)
	if entityErr != nil {
		err = entityErr
		return
	}

	queryVal, queryErr := s.querySingleModel(entityModel, entityFilter, 0)
	if queryErr != nil {
		err = queryErr
		return
	}

	if queryVal != nil {
		ret = queryVal
		return
	}

	err = fmt.Errorf("not exist model,name:%s,pkgPath:%s", entityModel.GetName(), entityModel.GetPkgPath())
	return
}
