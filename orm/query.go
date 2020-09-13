package orm

import (
	"fmt"
	"reflect"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

func (s *Orm) querySingle(modelInfo model.Model) (err error) {
	func() {
		builder := builder.NewBuilder(modelInfo, s.modelProvider)
		sql, sqlErr := builder.BuildQuery()
		if sqlErr != nil {
			err = sqlErr
			log.Errorf("build query failed, err:%s", err.Error())
			return
		}

		err = s.executor.Query(sql)
		if err != nil {
			return
		}

		defer s.executor.Finish()
		if !s.executor.Next() {
			err = fmt.Errorf("query %s failed, no found object", modelInfo.GetName())
			return
		}

		items, itemErr := s.getModelItems(modelInfo, builder)
		if itemErr != nil {
			err = itemErr
			return
		}

		err = s.executor.GetField(items...)
		if err != nil {
			return
		}

		idx := 0
		for _, item := range modelInfo.GetFields() {
			fType := item.GetType()
			depend := fType.Depend()
			if depend != nil && !util.IsBasicType(depend.GetValue()) {
				continue
			}

			err = item.UpdateValue(reflect.ValueOf(items[idx]).Elem())
			if err != nil {
				return
			}
			idx++
		}
	}()
	if err != nil {
		return
	}

	for _, item := range modelInfo.GetFields() {
		fType := item.GetType()
		fValue := item.GetValue()
		depend := fType.Depend()
		if depend == nil || util.IsBasicType(depend.GetValue()) || fValue.IsNil() {
			continue
		}

		itemVal, itemErr := s.queryRelation(modelInfo, item)
		if itemErr != nil {
			err = itemErr
			log.Errorf("queryRelation failed, modelName:%s, fieldName:%s, err:%s", modelInfo.GetName(), item.GetName(), err.Error())
			return
		}

		itemErr = item.UpdateValue(itemVal)
		if itemErr != nil {
			err = itemErr
			log.Errorf("UpdateFieldValue failed, modelName:%s, fieldName:%s, err:%s", modelInfo.GetName(), item.GetName(), err.Error())
			return
		}
	}

	return
}

func (s *Orm) queryRelation(modelInfo model.Model, fieldInfo model.Field) (ret reflect.Value, err error) {
	fieldType := fieldInfo.GetType()
	fieldModel, fieldErr := s.modelProvider.GetTypeModel(fieldType)
	if fieldErr != nil || fieldModel == nil {
		err = fieldErr
		if err == nil {
			err = fmt.Errorf("can't find typeModel")
		}

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

	var values []int64
	func() {
		err = s.executor.Query(relationSQL)
		if err != nil {
			return
		}

		defer s.executor.Finish()
		for s.executor.Next() {
			v := int64(0)
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

	if util.IsStructType(fieldType.GetValue()) {
		if len(values) > 0 {
			relationVal := reflect.Indirect(fieldModel.Interface())
			relationInfo, relationErr := s.modelProvider.GetValueModel(relationVal)
			if relationErr != nil {
				err = relationErr
				log.Errorf("GetValueModel failed, fieldName:%s, err:%s", fieldInfo.GetName(), err.Error())
				return
			}

			pkField := relationInfo.GetPrimaryField()
			err = pkField.UpdateValue(reflect.ValueOf(values[0]))
			if err != nil {
				log.Errorf("UpdateFieldValue pkField failed, fieldName:%s, err:%s", fieldInfo.GetName(), err.Error())
				return
			}

			err = s.querySingle(relationInfo)
			if err != nil {
				log.Errorf("querySingle for struct failed, fieldName:%s, err:%s", fieldInfo.GetName(), err.Error())
				return
			}

			ret = relationVal
			if fieldType.IsPtrType() {
				ret = ret.Addr()
			}
		}
	} else if util.IsSliceType(fieldType.GetValue()) {
		dependType := fieldType.Depend()
		relationVal := reflect.Indirect(fieldType.Interface())
		for _, item := range values {
			itemVal := fieldModel.Interface()
			itemInfo, itemErr := s.modelProvider.GetValueModel(itemVal)
			if itemErr != nil {
				log.Errorf("GetValueModel failed, err:%s", itemErr.Error())
				err = itemErr
				return
			}

			pkField := itemInfo.GetPrimaryField()
			err = pkField.UpdateValue(reflect.ValueOf(item))
			if err != nil {
				log.Errorf("UpdateValue failed, err:%s", err.Error())
				return
			}

			err = s.querySingle(itemInfo)
			if err != nil {
				log.Errorf("querySingle for slice failed, fieldName:%s, err:%s", fieldInfo.GetName(), err.Error())
				return
			}

			if dependType.IsPtrType() {
				itemVal = itemVal.Addr()
			}

			relationVal, err = s.modelProvider.AppendSliceValue(relationVal, itemVal)
			if err != nil {
				log.Errorf("append slice value failed, err:%s", err.Error())
				return
			}
		}

		ret = relationVal
		if fieldType.IsPtrType() {
			ret = ret.Addr()
		}
	}

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

	err = s.querySingle(entityModel)
	if err != nil {
		log.Errorf("querySingle failed, modelName:%s, err:%s", entityModel.GetName(), err.Error())
		return
	}

	return
}
