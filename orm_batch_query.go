package orm

import (
	"fmt"
	"log"
	"reflect"

	"muidea.com/magicOrm/builder"
	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/util"
)

func (s *orm) queryBatch(modelInfo model.Model, sliceValue reflect.Value, filter model.Filter) (ret reflect.Value, err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	sql, sqlErr := builder.BuildBatchQuery(filter)
	if sqlErr != nil {
		err = sqlErr
		return
	}

	s.executor.Query(sql)
	defer s.executor.Finish()
	for s.executor.Next() {
		newVal := modelInfo.Interface()
		newModelInfo, newErr := s.modelProvider.GetValueModel(newVal)
		if newErr != nil {
			err = newErr
			return
		}

		items := []interface{}{}
		fields := newModelInfo.GetFields()
		for _, field := range fields {
			fType := field.GetType()
			dependModel, dependErr := s.modelProvider.GetTypeModel(fType)
			if dependErr != nil {
				err = dependErr
				return
			}
			if dependModel != nil {
				continue
			}

			fieldVal, fieldErr := util.GetBasicTypeInitValue(fType.GetValue())
			if fieldErr != nil {
				err = fieldErr
				return
			}
			items = append(items, fieldVal)
		}
		s.executor.GetField(items...)

		idx := 0
		for _, field := range fields {
			fType := field.GetType()
			dependModel, dependErr := s.modelProvider.GetTypeModel(fType)
			if dependErr != nil {
				err = dependErr
				return
			}
			if dependModel != nil {
				continue
			}

			v := items[idx]
			err = field.UpdateValue(reflect.Indirect(reflect.ValueOf(v)))
			if err != nil {
				return
			}

			idx++
		}

		sliceValue = reflect.Append(sliceValue, newVal)
	}

	ret = sliceValue

	return
}

func (s *orm) BatchQuery(sliceObj interface{}, filter model.Filter) (err error) {
	objValue := reflect.ValueOf(sliceObj)
	if objValue.Kind() != reflect.Ptr {
		err = fmt.Errorf("illegal obj type. must be a slice ptr")
		return
	}

	modelInfo, modelErr := s.modelProvider.GetObjectModel(objValue)
	if modelErr != nil {
		err = modelErr
		log.Printf("GetTypeModel failed, err:%s", err.Error())
		return
	}

	objValue = reflect.Indirect(objValue)
	queryValues, queryErr := s.queryBatch(modelInfo, objValue, filter)
	if queryErr != nil {
		err = queryErr
		return
	}

	objValue.Set(queryValues)

	return
}
