package orm

import (
	"fmt"
	"log"
	"reflect"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

func (s *Orm) querySingle(modelInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	sql, err := builder.BuildQuery()
	if err != nil {
		log.Printf("build query failed, err:%s", err.Error())
		return err
	}

	s.executor.Query(sql)
	defer s.executor.Finish()
	if !s.executor.Next() {
		err = fmt.Errorf("query %s failed, no found object", modelInfo.GetName())
		return
	}

	items, itemErr := s.getItems(modelInfo)
	if itemErr != nil {
		err = itemErr
		return
	}

	s.executor.GetField(items...)

	idx := 0
	for _, item := range modelInfo.GetFields() {
		fType := item.GetType()
		depend := fType.Depend()
		if depend != nil && !util.IsBasicType(depend.GetValue()) {
			continue
		}

		v := items[idx]
		err = item.UpdateValue(reflect.Indirect(reflect.ValueOf(v)))
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
		log.Printf("GetTypeModel failed, type:%s, err:%s", fType.GetType().String(), err.Error())
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
		log.Printf("BuildQueryRelation failed, fieldName:%s, err:%s", fieldInfo.GetName(), err.Error())
		return
	}

	values := []int64{}
	func() {
		s.executor.Query(relationSQL)
		defer s.executor.Finish()
		for s.executor.Next() {
			v := int64(0)
			s.executor.GetField(&v)
			values = append(values, v)
		}
	}()

	if util.IsStructType(fType.GetValue()) {
		if len(values) > 0 {
			relationVal := fieldModel.Interface()
			relationInfo, relationErr := s.modelProvider.GetValueModel(relationVal)
			if relationErr != nil {
				log.Printf("GetValueModel failed, fieldName:%s, err:%s", fieldInfo.GetName(), err.Error())
				err = relationErr
				return
			}

			pkField := relationInfo.GetPrimaryField()
			err = pkField.UpdateValue(reflect.ValueOf(values[0]))
			if err != nil {
				log.Printf("UpdateFieldValue pkField failed, fieldName:%s, err:%s", fieldInfo.GetName(), err.Error())
				return
			}

			err = s.querySingle(relationInfo)
			if err != nil {
				log.Printf("querySingle for struct failed, fieldName:%s, err:%s", fieldInfo.GetName(), err.Error())
				return
			}

			for _, item := range relationInfo.GetFields() {
				itemVal, itemErr := s.queryRelation(relationInfo, item)
				if itemErr != nil {
					//err = itemErr
					log.Printf("queryRelation failed, modelName:%s, field:%s, err:%s", relationInfo.GetName(), item.GetName(), err.Error())
					//return
					continue
				}

				if util.IsNil(itemVal) {
					continue
				}

				err = relationInfo.UpdateFieldValue(item.GetName(), itemVal)
				if err != nil {
					log.Printf("UpdateFieldValue failed, fieldName:%s, err:%s", item.GetName(), err.Error())
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
				log.Printf("GetValueModel faield, err:%s", itemErr.Error())
				err = itemErr
				return
			}

			pkField := itemInfo.GetPrimaryField()
			err = pkField.UpdateValue(reflect.ValueOf(item))
			if err != nil {
				log.Printf("UpdateValue failed, err:%s", err.Error())
				return
			}

			err = s.querySingle(itemInfo)
			if err != nil {
				log.Printf("querySingle for slice failed, fieldName:%s, err:%s", fieldInfo.GetName(), err.Error())
				return
			}

			for _, item := range itemInfo.GetFields() {
				subVal, subErr := s.queryRelation(itemInfo, item)
				if subErr != nil {
					//err = subErr
					log.Printf("queryRelation failed, modelName:%s, field:%s, err:%s", itemInfo.GetName(), item.GetName(), err.Error())
					//return
					continue
				}
				if util.IsNil(subVal) {
					continue
				}

				err = itemInfo.UpdateFieldValue(item.GetName(), subVal)
				if err != nil {
					log.Printf("UpdateFieldValue failed, fieldName:%s, err:%s", item.GetName(), err.Error())
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
	entityVal := reflect.ValueOf(entity)
	modelInfo, modelErr := s.modelProvider.GetValueModel(entityVal)
	if modelErr != nil {
		err = modelErr
		log.Printf("GetValueModel failed, err:%s", err.Error())
		return
	}

	err = s.querySingle(modelInfo)
	if err != nil {
		log.Printf("querySingle failed, modelName:%s, err:%s", modelInfo.GetName(), err.Error())
		return
	}

	for _, item := range modelInfo.GetFields() {
		itemVal, itemErr := s.queryRelation(modelInfo, item)
		if itemErr != nil {
			err = itemErr
			log.Printf("queryRelation failed, modelName:%s, field:%s, err:%s", modelInfo.GetName(), item.GetName(), err.Error())
			return
		}
		if util.IsNil(itemVal) {
			continue
		}

		err = modelInfo.UpdateFieldValue(item.GetName(), itemVal)
		if err != nil {
			log.Printf("UpdateFieldValue failed, fieldName:%s, err:%s", item.GetName(), err.Error())
			return
		}
	}

	return
}
