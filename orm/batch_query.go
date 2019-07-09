package orm

import (
	"fmt"
	"log"
	"reflect"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

type resultItems []interface{}

func (s *Orm) queryBatch(modelInfo model.Model, sliceValue reflect.Value, filter model.Filter) (err error) {
	var maskModel model.Model
	if filter != nil {
		maskModel, err = filter.MaskModel()
		if err != nil {
			log.Printf("Get MaskModel failed, err:%s", err.Error())
			return
		}
		if maskModel != nil {
			if maskModel.GetName() != modelInfo.GetName() || maskModel.GetPkgPath() != modelInfo.GetPkgPath() {
				err = fmt.Errorf("illegal value mask")
				return
			}
		}
	}

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

	resultSlice := reflect.MakeSlice(sliceValue.Type(), 0, 0)
	for idx := 0; idx < len(queryList); idx++ {
		var newVal reflect.Value
		if maskModel != nil {
			newVal = maskModel.Interface()
		} else {
			newVal = modelInfo.Interface()
		}
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
				itemVal, itemErr := s.queryRelation(newModelInfo, field)
				if itemErr != nil {
					err = itemErr
					log.Printf("queryRelation failed, err:%s", err.Error())
					return
				}
				if util.IsNil(itemVal) {
					continue
				}

				newModelInfo.UpdateFieldValue(field.GetName(), itemVal)
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

		resultSlice = reflect.Append(resultSlice, newVal)
	}

	sliceValue.Set(resultSlice)

	return
}

// BatchQuery batch query
func (s *Orm) BatchQuery(sliceEntity interface{}, filter model.Filter) (err error) {
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

	queryErr := s.queryBatch(sliceModel, sliceVal, filter)
	if queryErr != nil {
		err = queryErr
		log.Printf("queryBatch failed, err:%s", err.Error())
		return
	}

	return
}
