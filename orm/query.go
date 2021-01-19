package orm

import (
	"fmt"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

type resultItems []interface{}

func (s *impl) querySingle(vModel model.Model, filter model.Filter) (ret model.Model, err error) {
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
			err = fmt.Errorf("query %s failed, no found object", vModel.GetName())
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
	if err != nil {
		return
	}

	modelVal, modelErr := s.assignSingleModel(vModel, queryValue)
	if modelErr != nil {
		err = modelErr
		return
	}

	ret = modelVal
	return
}

func (s *impl) assignSingleModel(modelVal model.Model, queryVal resultItems) (ret model.Model, err error) {
	offset := 0
	for _, field := range modelVal.GetFields() {
		fType := field.GetType()
		fValue := field.GetValue()
		if !fType.IsBasic() && !fValue.IsNil() {
			itemVal, itemErr := s.queryRelation(modelVal, field)
			if itemErr != nil {
				continue
			}

			if itemVal != nil {
				err = field.SetValue(itemVal)
				if err != nil {
					return
				}
			}

			//offset++
			continue
		}

		if !fValue.IsNil() {
			fVal, fErr := fType.Interface(s.stripSlashes(fType, queryVal[offset]))
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

func (s *impl) queryRelationSingle(id int, vModel model.Model) (ret model.Model, err error) {
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

	queryVal, queryErr := s.querySingle(relationModel, relationFilter)
	if queryErr != nil {
		err = queryErr
		return
	}

	ret = queryVal
	return
}

func (s *impl) queryRelationSlice(ids []int, vModel model.Model) (ret []model.Model, err error) {
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

		queryVal, queryErr := s.querySingle(relationModel, relationFilter)
		if queryErr != nil {
			err = queryErr
			return
		}

		sliceVal = append(sliceVal, queryVal)
	}

	ret = sliceVal
	return
}

func (s *impl) queryRelation(modelInfo model.Model, fieldInfo model.Field) (ret model.Value, err error) {
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
		return
	}

	fieldValue, _ := fieldType.Interface(nil)
	if util.IsStructType(fieldType.GetValue()) {
		queryVal, queryErr := s.queryRelationSingle(values[0], fieldModel)
		if queryErr != nil {
			err = queryErr
			return
		}

		modelVal, modelErr := s.modelProvider.GetEntityValue(queryVal.Interface(true))
		if modelErr != nil {
			err = modelErr
			return
		}

		fieldValue.Set(modelVal.Get())
	} else if util.IsSliceType(fieldType.GetValue()) {
		queryVal, queryErr := s.queryRelationSlice(values, fieldModel)
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

	if fieldType.IsPtrType() {
		fieldValue = fieldValue.Addr()
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

	queryVal, queryErr := s.querySingle(entityModel, entityFilter)
	if queryErr != nil {
		err = queryErr
		return
	}

	ret = queryVal
	return
}
