package orm

import (
	"fmt"
	"log"
	"reflect"

	"muidea.com/magicOrm/builder"
	"muidea.com/magicOrm/filter"
	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/util"
)

func (s *orm) queryBatch(structInfo model.Model, sliceValue reflect.Value, filter filter.Filter) (ret reflect.Value, err error) {
	builder := builder.NewBuilder(structInfo)
	sql, sqlErr := builder.BuildBatchQuery(filter)
	if sqlErr != nil {
		err = sqlErr
		return
	}

	s.executor.Query(sql)
	for s.executor.Next() {
		newVal := structInfo.Interface()
		newStructInfo, newErr := model.GetStructValue(newVal, s.modelInfoCache)
		if newErr != nil {
			err = newErr
			return
		}

		items := []interface{}{}
		fields := newStructInfo.GetFields()
		for _, val := range *fields {
			fType := val.GetFieldType()

			dependType, _ := fType.Depend()
			if dependType != nil {
				continue
			}

			v := util.GetBasicTypeInitValue(fType.Value())
			items = append(items, v)
		}
		s.executor.GetField(items...)

		idx := 0
		for _, val := range *fields {
			fType := val.GetFieldType()

			dependType, _ := fType.Depend()
			if dependType != nil {
				continue
			}

			v := items[idx]
			err = val.SetFieldValue(reflect.Indirect(reflect.ValueOf(v)))
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

func (s *orm) BatchQuery(sliceObj interface{}, filter filter.Filter) (err error) {
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
	structInfo, structErr := model.GetStructInfo(rawType, s.modelInfoCache)
	if structErr != nil {
		err = structErr
		log.Printf("GetStructInfo failed, err:%s", err.Error())
		return
	}

	objValue := reflect.ValueOf(sliceObj)
	objValue = reflect.Indirect(objValue)
	queryValues, queryErr := s.queryBatch(structInfo, objValue, filter)
	if queryErr != nil {
		err = queryErr
		return
	}

	objValue.Set(queryValues)

	return
}
