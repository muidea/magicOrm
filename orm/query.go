package orm

import (
	"fmt"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

type resultItems []interface{}

func (s *impl) querySingle(vModel model.Model, filter model.Filter, deepLevel int) (ret model.Model, err error) {
	var queryValue resultItems
	func() {
		builder := builder.NewBuilder(vModel, s.modelProvider)
		sqlStr, sqlErr := builder.BuildQuery(filter)
		if sqlErr != nil {
			err = sqlErr
			return
		}

		err = s.executor.Query(sqlStr)
		if err != nil {
			return
		}

		defer s.executor.Finish()
		if !s.executor.Next() {
			return
		}

		items, itemErr := s.getInitializeValue(vModel, builder)
		if itemErr != nil {
			err = itemErr
			return
		}

		err = s.executor.GetField(items...)
		if err != nil {
			return
		}

		queryValue = items
	}()
	if err != nil || len(queryValue) == 0 {
		return
	}

	modelVal, modelErr := s.assignSingleModel(vModel, queryValue, deepLevel)
	if modelErr != nil {
		err = modelErr
		return
	}

	ret = modelVal
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

func (s *impl) queryRelationSingle(id int, vModel model.Model, deepLevel int) (ret model.Model, err error) {
	relationModel := vModel.Copy()
	relationVal, relationErr := s.modelProvider.GetEntityValue(id)
	if relationErr != nil {
		err = fmt.Errorf("GetEntityValue failed, err:%s", relationErr)
		return
	}

	pkField := relationModel.GetPrimaryField()
	pkField.SetValue(relationVal)
	relationFilter, relationErr := s.getFieldFilter(pkField)
	if relationErr != nil {
		err = fmt.Errorf("GetEntityValue failed, err:%s", relationErr)
		return
	}

	queryVal, queryErr := s.querySingle(relationModel, relationFilter, deepLevel)
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
		relationFilter, relationErr := s.getFieldFilter(pkField)
		if relationErr != nil {
			err = fmt.Errorf("GetEntityValue failed, err:%s", relationErr)
			return
		}

		queryVal, queryErr := s.querySingle(relationModel, relationFilter, deepLevel)
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

func (s *impl) queryRelation(modelInfo model.Model, fieldInfo model.Field, deepLevel int) (ret model.Value, err error) {
	if deepLevel > 1 {
		return
	}

	fieldType := fieldInfo.GetType()
	fieldModel, fieldErr := s.modelProvider.GetTypeModel(fieldType)
	if fieldErr != nil {
		err = fieldErr
		return
	}

	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	relationSQL, relationErr := builder.BuildQueryRelation(fieldInfo.GetName(), fieldModel)
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

	if err != nil || len(values) == 0 {
		ret, err = fieldType.Interface()
		return
	}

	fieldValue, _ := fieldType.Interface()
	if util.IsStructType(fieldType.GetValue()) {
		queryVal, queryErr := s.queryRelationSingle(values[0], fieldModel, deepLevel+1)
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

	if util.IsSliceType(fieldType.GetValue()) {
		queryVal, queryErr := s.queryRelationSlice(values, fieldModel, deepLevel)
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
	}
	ret = fieldValue

	return
}

// Query query
func (s *impl) Query(entityModel model.Model) (ret model.Model, err error) {
	entityFilter, entityErr := s.getModelFilter(entityModel)
	if entityErr != nil {
		err = entityErr
		return
	}

	queryVal, queryErr := s.querySingle(entityModel, entityFilter, 0)
	if queryErr != nil {
		err = queryErr
		return
	}

	if queryVal != nil {
		ret = queryVal
		return
	}

	err = fmt.Errorf("not exist model")
	return
}
