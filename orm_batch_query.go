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

			dependModel, dependErr := s.modelProvider.GetTypeModel(fType.GetType())
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
			fValue := field.GetValue()

			dependModel, dependErr := s.modelProvider.GetTypeModel(fType.GetType())
			if dependErr != nil {
				err = dependErr
				return
			}
			if dependModel != nil {
				continue
			}

			v := items[idx]
			err = fValue.Set(reflect.Indirect(reflect.ValueOf(v)))
			if err != nil {
				return
			}

			idx++
		}

		sliceValue = reflect.Append(sliceValue, newVal.Elem())
	}

	ret = sliceValue

	return
}

func (s *orm) BatchQuery(sliceObj interface{}, filter model.Filter) (err error) {
	objType := reflect.TypeOf(sliceObj)
	if objType.Kind() != reflect.Ptr {
		err = fmt.Errorf("illegal obj type. must be a slice ptr")
		return
	}

	rawType := objType.Elem()
	if rawType.Kind() != reflect.Slice {
		err = fmt.Errorf("illegal obj type. must be a slice ptr")
		return
	}
	rawType = rawType.Elem()
	modelInfo, modelErr := s.modelProvider.GetTypeModel(rawType)
	if modelErr != nil {
		err = modelErr
		log.Printf("GetTypeModel failed, err:%s", err.Error())
		return
	}

	objValue := reflect.ValueOf(sliceObj)
	objValue = reflect.Indirect(objValue)
	queryValues, queryErr := s.queryBatch(modelInfo, objValue, filter)
	if queryErr != nil {
		err = queryErr
		return
	}

	objValue.Set(queryValues)

	return
}
