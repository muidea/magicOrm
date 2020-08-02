package orm

import (
	"fmt"
	"reflect"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/helper"
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

		itemVal := items[idx]
		typeVal := fType.Interface()
		typeVal, err = helper.AssignValue(reflect.ValueOf(itemVal).Elem(), typeVal)
		if err != nil {
			log.Errorf("assignValue failed, name:%s", item.GetName())
			return err
		}

		err = item.UpdateValue(typeVal)
		if err != nil {
			return err
		}
		idx++
	}

	return
}

func (s *Orm) queryRelation(modelInfo model.Model, fieldInfo model.Field) (ret reflect.Value, err error) {
	fType := fieldInfo.GetType()
	fieldModel, fieldErr := s.modelProvider.GetTypeModel(fType)
	if fieldErr != nil {
		err = fieldErr
		log.Errorf("GetTypeModel failed, type:%s, err:%s", fType.GetName(), err.Error())
		return
	}
	if fieldModel == nil {
		return
	}

	fValue := fieldInfo.GetValue()
	if fValue.IsNil() {
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

	if util.IsStructType(fType.GetValue()) {
		if len(values) > 0 {
			relationVal := fieldModel.Interface()
			relationInfo, relationErr := s.modelProvider.GetValueModel(relationVal)
			if relationErr != nil {
				err = relationErr
				log.Errorf("GetValueModel failed, fieldName:%s, err:%s", fieldInfo.GetName(), err.Error())
				return
			}

			pkField := relationInfo.GetPrimaryField()
			qVal := reflect.ValueOf(values[0])
			fVal := pkField.GetType().Interface()
			fVal, err = helper.AssignValue(qVal, fVal)
			if err != nil {
				log.Errorf("assign pk field failed, err:%s", err.Error())
				return
			}

			err = pkField.UpdateValue(fVal)
			if err != nil {
				log.Errorf("UpdateFieldValue pkField failed, fieldName:%s, err:%s", fieldInfo.GetName(), err.Error())
				return
			}

			err = s.querySingle(relationInfo)
			if err != nil {
				log.Errorf("querySingle for struct failed, fieldName:%s, err:%s", fieldInfo.GetName(), err.Error())
				return
			}

			for _, item := range relationInfo.GetFields() {
				itemVal, itemErr := s.queryRelation(relationInfo, item)
				if itemErr != nil {
					//err = itemErr
					log.Errorf("queryRelation failed, modelName:%s, field:%s, err:%s", relationInfo.GetName(), item.GetName(), itemErr.Error())
					//return
					continue
				}

				err = item.UpdateValue(itemVal)
				if err != nil {
					log.Errorf("UpdateFieldValue failed, fieldName:%s, err:%s", item.GetName(), err.Error())
					return
				}
			}

			ret = relationVal
		}
	} else if util.IsSliceType(fType.GetValue()) {
		relationVal := reflect.Indirect(fType.Interface())
		for _, item := range values {
			itemVal := reflect.Indirect(fieldModel.Interface())
			itemInfo, itemErr := s.modelProvider.GetValueModel(itemVal)
			if itemErr != nil {
				log.Errorf("GetValueModel faield, err:%s", itemErr.Error())
				err = itemErr
				return
			}

			pkField := itemInfo.GetPrimaryField()
			qVal := reflect.ValueOf(item)
			fVal := pkField.GetType().Interface()
			fVal, err = helper.AssignValue(qVal, fVal)
			if err != nil {
				log.Errorf("assign pk field failed, err:%s", err.Error())
				return
			}

			err = pkField.UpdateValue(fVal)
			if err != nil {
				log.Errorf("UpdateValue failed, err:%s", err.Error())
				return
			}

			err = s.querySingle(itemInfo)
			if err != nil {
				log.Errorf("querySingle for slice failed, fieldName:%s, err:%s", fieldInfo.GetName(), err.Error())
				return
			}

			for _, item := range itemInfo.GetFields() {
				subVal, subErr := s.queryRelation(itemInfo, item)
				if subErr != nil {
					//err = subErr
					log.Errorf("queryRelation failed, modelName:%s, field:%s, err:%s", itemInfo.GetName(), item.GetName(), subErr.Error())
					//return
					continue
				}

				err = item.UpdateValue(subVal)
				if err != nil {
					log.Errorf("UpdateFieldValue failed, fieldName:%s, err:%s", item.GetName(), err.Error())
					return
				}
			}

			if fieldModel.IsPtrModel() {
				itemVal = itemVal.Addr()
			}

			relationVal = reflect.Append(relationVal, itemVal)
		}

		ret = relationVal
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

	for _, item := range modelInfo.GetFields() {
		fType := item.GetType()
		depend := fType.Depend()
		if depend == nil {
			continue
		}

		itemVal, itemErr := s.queryRelation(modelInfo, item)
		if itemErr != nil {
			log.Errorf("queryRelation failed, modelName:%s, field:%s, err:%s", modelInfo.GetName(), item.GetName(), itemErr.Error())
			continue
		}

		err = item.UpdateValue(itemVal)
		if err != nil {
			log.Errorf("UpdateFieldValue failed, fieldName:%s, err:%s", item.GetName(), err.Error())
			return
		}
	}

	return
}
