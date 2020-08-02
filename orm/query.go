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
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	sql, err := builder.BuildQuery()
	if err != nil {
		log.Errorf("build query failed, err:%s", err.Error())
		return err
	}

	log.Infof("sql:%s", sql)

	err = s.executor.Query(sql)
	if err != nil {
		return
	}

	defer s.executor.Finish()
	if !s.executor.Next() {
		err = fmt.Errorf("query %s failed, no found object", modelInfo.GetName())
		return
	}

	items, itemErr := s.getModelItems(modelInfo)
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
		if depend != nil {
			continue
		}

		err = item.UpdateValue(reflect.ValueOf(items[idx]).Elem())
		if err != nil {
			return err
		}
		idx++
	}

	for _, item := range modelInfo.GetFields() {
		fType := item.GetType()
		fValue := item.GetValue()
		depend := fType.Depend()
		if depend == nil || fValue.IsNil() {
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
			log.Errorf("type:%v", relationVal.Type().String())
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
				if relationVal.CanAddr() {
					ret = relationVal.Addr()
					return
				}

				ret = reflect.New(relationVal.Type()).Elem()
				ret.Set(relationVal)
				ret = ret.Addr()
			}
		}
	} else if util.IsSliceType(fieldType.GetValue()) {
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

			relationVal = reflect.Append(relationVal, itemVal)
		}

		ret = relationVal
		if fieldType.IsPtrType() {
			if relationVal.CanAddr() {
				ret = relationVal.Addr()
				return
			}

			ret = reflect.New(relationVal.Type()).Elem()
			ret.Set(relationVal)
			ret = ret.Addr()
		}
	}

	return
}

// Query query
func (s *Orm) Query(entity interface{}) (err error) {
	entityVal := reflect.ValueOf(entity).Elem()
	modelInfo, modelErr := s.modelProvider.GetValueModel(entityVal)
	if modelErr != nil {
		err = modelErr
		log.Errorf("GetValueModel failed, err:%s", err.Error())
		return
	}

	err = s.querySingle(modelInfo)
	if err != nil {
		log.Errorf("querySingle failed, modelName:%s, err:%s", modelInfo.GetName(), err.Error())
		return
	}

	return
}
