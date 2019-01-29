package orm

import (
	"fmt"
	"log"
	"reflect"

	"muidea.com/magicOrm/builder"
	"muidea.com/magicOrm/local"
	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/util"
)

func (s *orm) queryBatch(modelInfo model.Model, sliceValue reflect.Value, filter model.Filter) (ret reflect.Value, err error) {
	builder := builder.NewBuilder(modelInfo)
	sql, sqlErr := builder.BuildBatchQuery(filter, s.modelInfoCache)
	if sqlErr != nil {
		err = sqlErr
		return
	}

	s.executor.Query(sql)
	for s.executor.Next() {
		newVal := modelInfo.Interface()
		newStructInfo, newErr := local.GetValueModel(newVal, s.modelInfoCache)
		if newErr != nil {
			err = newErr
			return
		}

		items := []interface{}{}
		fields := newStructInfo.GetFields()
		for _, val := range *fields {
			fType := val.GetType()

			dependType := fType.Depend()
			if dependType != nil {
				continue
			}

			v := util.GetBasicTypeInitValue(fType.Value())
			items = append(items, v)
		}
		s.executor.GetField(items...)

		idx := 0
		for _, val := range *fields {
			fType := val.GetType()

			dependType := fType.Depend()
			if dependType != nil {
				continue
			}

			v := items[idx]
			err = val.SetValue(reflect.Indirect(reflect.ValueOf(v)))
			if err != nil {
				return
			}

			idx++
		}

		sliceValue = reflect.Append(sliceValue, newVal.Elem())
	}

	defer s.executor.Finish()

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
	modelInfo, structErr := local.GetTypeModel(rawType, s.modelInfoCache)
	if structErr != nil {
		err = structErr
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
