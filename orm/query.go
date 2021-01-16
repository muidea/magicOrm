package orm

import (
	"fmt"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

type resultItems []interface{}

func (s *Orm) querySingle(vModel model.Model, filter model.Filter) (ret model.Value, err error) {
	var queryValue resultItems
	func() {
		builder := builder.NewBuilder(vModel, s.modelProvider)
		sqlStr, sqlErr := builder.BuildQuery(filter)
		if sqlErr != nil {
			err = sqlErr
			log.Errorf("build query failed, err:%s", err.Error())
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
		log.Errorf("assignSingle model failed, err:%s", err.Error())
		return
	}

	ret = modelVal
	return
}

func (s *Orm) assignSingleModel(modelVal model.Model, queryVal resultItems) (ret model.Value, err error) {
	offset := 0
	for _, field := range modelVal.GetFields() {
		fType := field.GetType()
		fValue := field.GetValue()
		if !fType.IsBasic() && !fValue.IsNil() {
			itemVal, itemErr := s.queryRelation(modelVal, field)
			if itemErr != nil {
				log.Errorf("queryRelation failed, err:%s", itemErr.Error())
				continue
			}

			if itemVal != nil {
				err = field.SetValue(itemVal)
				if err != nil {
					log.Errorf("SetValue failed, err:%s", err.Error())
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
				log.Errorf("Interface failed, err:%s", err.Error())
				return
			}
			err = field.SetValue(fVal)
			if err != nil {
				log.Errorf("SetValue failed, err:%s", err.Error())
				return
			}
		}

		offset++
	}

	ret = modelVal.Interface()
	return
}

func (s *Orm) queryRelationSingle(id int, vModel model.Model) (ret model.Value, err error) {
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
		log.Errorf("querySingle for struct failed, err:%s", err.Error())
		return
	}

	ret = queryVal
	return
}

func (s *Orm) queryRelationSlice(ids []int, vModel model.Model, sliceVal model.Value) (ret model.Value, err error) {
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
			log.Errorf("querySingle for slice failed, err:%s", err.Error())
			return
		}

		sliceVal, err = s.modelProvider.AppendSliceValue(sliceVal, queryVal)
		if err != nil {
			log.Errorf("append slice value failed, err:%s", err.Error())
			return
		}
	}

	ret = sliceVal
	return
}

func (s *Orm) queryRelation(modelInfo model.Model, fieldInfo model.Field) (ret model.Value, err error) {
	fieldType := fieldInfo.GetType()
	fieldModel, fieldErr := s.modelProvider.GetTypeModel(fieldType)
	if fieldErr != nil {
		err = fieldErr
		log.Errorf("GetTypeModel failed, type:%s, err:%s", fieldType.GetName(), err.Error())
		return
	}

	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	relationSQL, relationErr := builder.BuildQueryRelation(fieldInfo.GetName(), fieldModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("BuildQueryRelation failed, fieldName:%s, err:%s", fieldInfo.GetName(), err.Error())
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
			log.Errorf("queryRelationSingle failed, fieldName:%s, err:%s", fieldInfo.GetName(), err.Error())
			return
		}

		fieldValue.Set(queryVal.Get())
	} else if util.IsSliceType(fieldType.GetValue()) {
		queryVal, queryErr := s.queryRelationSlice(values, fieldModel, fieldValue)
		if queryErr != nil {
			err = queryErr
			log.Errorf("queryRelationSlice failed, fieldName:%s, err:%s", fieldInfo.GetName(), err.Error())
			return
		}

		fieldValue.Set(queryVal.Get())
	}

	if fieldType.IsPtrType() {
		fieldValue = fieldValue.Addr()
	}

	ret = fieldValue
	return
}

// Query query
func (s *Orm) Query(entity interface{}) (err error) {
	entityModel, entityErr := s.modelProvider.GetEntityModel(entity)
	if entityErr != nil {
		err = entityErr
		log.Errorf("GetEntityModel failed, err:%s", err.Error())
		return
	}

	entityVal, entityErr := s.modelProvider.GetEntityValue(entity)
	if entityErr != nil {
		err = entityErr
		log.Errorf("GetEntityValue failed, err:%s", err.Error())
		return
	}

	entityFilter, entityErr := s.getModelFilter(entityModel)
	if entityErr != nil {
		err = entityErr
		log.Errorf("getModelFilter failed, err:%s", err.Error())
		return
	}

	queryVal, queryErr := s.querySingle(entityModel, entityFilter)
	if queryErr != nil {
		err = queryErr
		log.Errorf("querySingle failed, modelName:%s, err:%s", entityModel.GetName(), err.Error())
		return
	}

	entityVal.Set(queryVal.Get())
	return
}
