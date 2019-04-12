package orm

import (
	"fmt"
	"log"
	"reflect"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

type resultItems []interface{}

func (s *orm) queryBatch(modelInfo model.Model, sliceValue reflect.Value, filter model.Filter) (ret reflect.Value, err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	sql, sqlErr := builder.BuildBatchQuery(filter)
	if sqlErr != nil {
		err = sqlErr
		log.Printf("BuildBatchQuery failed, err:%s", err.Error())
		return
	}

	queryList := []resultItems{}
	s.executor.Query(sql)
	defer s.executor.Finish()
	for s.executor.Next() {
		modelItems, modelErr := s.getItems(modelInfo)
		if modelErr != nil {
			err = modelErr
			return
		}

		s.executor.GetField(modelItems...)

		queryList = append(queryList, modelItems)
	}

	for idx := 0; idx < len(queryList); idx++ {
		newVal := modelInfo.Interface()
		newModelInfo, newErr := s.modelProvider.GetValueModel(newVal)
		if newErr != nil {
			err = newErr
			log.Printf("GetValueModel failed, err:%s", err.Error())
			return
		}
		fields := newModelInfo.GetFields()
		offset := 0
		for _, field := range fields {
			fType := field.GetType()
			dependModel, dependErr := s.modelProvider.GetTypeModel(fType)
			if dependErr != nil {
				err = dependErr
				log.Printf("GetTypeModel failed, err:%s", err.Error())
				return
			}
			if dependModel != nil {
				err = s.queryRelation(newModelInfo, field)
				if err != nil {
					return
				}
				continue
			}

			v := queryList[idx][offset]
			err = field.UpdateValue(reflect.Indirect(reflect.ValueOf(v)))
			if err != nil {
				log.Printf("UpdateValue failed, err:%s", err.Error())
				return
			}

			offset++
		}

		sliceValue = reflect.Append(sliceValue, newVal)
	}

	ret = sliceValue

	return
}

func (s *orm) BatchQuery(sliceEntity interface{}, filter model.Filter) (ret interface{}, err error) {
	sliceEntityVal := reflect.ValueOf(sliceEntity)
	if sliceEntityVal.Kind() != reflect.Ptr {
		err = fmt.Errorf("illegal obj type. must be a slice ptr")
		return
	}

	sliceModel, sliceVal, modelErr := s.modelProvider.GetSliceValueModel(sliceEntityVal)
	if modelErr != nil {
		err = modelErr
		log.Printf("GetTypeModel failed, err:%s", err.Error())
		return
	}

	queryValues, queryErr := s.queryBatch(sliceModel, sliceVal, filter)
	if queryErr != nil {
		err = queryErr
		log.Printf("queryBatch failed, err:%s", err.Error())
		return
	}

	ret = queryValues.Interface()

	return
}
