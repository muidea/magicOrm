package orm

import (
	"fmt"
	"log"
	"reflect"

	"muidea.com/magicOrm/builder"
	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/util"
)

func (s *orm) querySingle(modelInfo model.Model) (err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	sql, err := builder.BuildQuery()
	if err != nil {
		return err
	}

	s.executor.Query(sql)
	if !s.executor.Next() {
		return fmt.Errorf("query %s failed, no found object", modelInfo.GetName())
	}
	defer s.executor.Finish()

	items, itemErr := s.getItems(modelInfo)
	if itemErr != nil {
		err = itemErr
		return
	}

	s.executor.GetField(items...)

	idx := 0
	for _, item := range modelInfo.GetFields() {
		fType := item.GetType()
		dependModel, dependErr := s.modelProvider.GetTypeModel(fType)
		if dependErr != nil {
			err = dependErr
			return
		}
		if dependModel != nil {
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

func (s *orm) queryRelation(modelInfo model.Model, fieldInfo model.Field) (err error) {
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
		return err
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
				return
			}

			err = s.querySingle(relationInfo)
			if err != nil {
				log.Printf("querySingle for struct failed, fieldName:%s, err:%s", fieldInfo.GetName(), err.Error())
				return
			}

			err = modelInfo.UpdateFieldValue(fieldInfo.GetName(), relationVal)
			if err != nil {
				log.Printf("UpdateFieldValue failed, fieldName:%s, err:%s", fieldInfo.GetName(), err.Error())
				return
			}
		}
	} else if util.IsSliceType(fType.GetValue()) {
		relationVal := reflect.Indirect(fType.Interface())
		for _, item := range values {
			itemVal := fieldModel.Interface()
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

			relationVal = reflect.Append(relationVal, itemVal)
		}

		err = modelInfo.UpdateFieldValue(fieldInfo.GetName(), relationVal)
		if err != nil {
			log.Printf("UpdateFieldValue failed, fieldName:%s, err:%s", fieldInfo.GetName(), err.Error())
			return
		}
	}

	return
}

func (s *orm) Query(obj interface{}) (err error) {
	modelInfo, modelErr := s.modelProvider.GetObjectModel(obj)
	if modelErr != nil {
		err = modelErr
		log.Printf("GetObjectModel failed, err:%s", err.Error())
		return
	}

	err = s.querySingle(modelInfo)
	if err != nil {
		log.Printf("querySingle failed, modelName:%s, err:%s", modelInfo.GetName(), err.Error())
		return
	}

	for _, item := range modelInfo.GetFields() {
		err = s.queryRelation(modelInfo, item)
		if err != nil {
			log.Printf("queryRelation failed, modelName:%s, field:%s, err:%s", modelInfo.GetName(), item.GetName(), err.Error())
			return
		}
	}

	return
}
