package orm

import (
	"fmt"
	"reflect"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/helper"
)

type resultItems []interface{}

func (s *Orm) queryBatch(modelInfo model.Model, sliceValue reflect.Value, filter model.Filter) (err error) {
	var maskModel model.Model
	if filter != nil {
		maskModel, err = filter.MaskModel()
		if err != nil {
			log.Errorf("Get MaskModel failed, err:%s", err.Error())
			return
		}

		if maskModel != nil {
			if maskModel.GetName() != modelInfo.GetName() || maskModel.GetPkgPath() != modelInfo.GetPkgPath() {
				err = fmt.Errorf("illegal value mask")
				return
			}
		}
	}
	if maskModel == nil {
		maskModel = modelInfo
	}

	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	sql, sqlErr := builder.BuildBatchQuery(filter)
	if sqlErr != nil {
		err = sqlErr
		log.Errorf("BuildBatchQuery failed, err:%s", err.Error())
		return
	}

	var queryList []resultItems
	err = s.executor.Query(sql)
	if err != nil {
		return
	}
	defer s.executor.Finish()
	for s.executor.Next() {
		modelItems, modelErr := s.getModelItems(maskModel)
		if modelErr != nil {
			err = modelErr
			return
		}

		err = s.executor.GetField(modelItems...)
		if err != nil {
			return
		}

		queryList = append(queryList, modelItems)
	}

	resultSlice := reflect.MakeSlice(sliceValue.Type(), 0, 0)
	for idx := 0; idx < len(queryList); idx++ {
		modelVal := maskModel.Interface()
		modelVal, err = s.assignSingleModel(modelVal, queryList[idx])
		if err != nil {
			log.Errorf("assignSingle model failed, err:%s", err.Error())
			return
		}

		resultSlice = reflect.Append(resultSlice, modelVal)
	}

	sliceValue.Set(resultSlice)

	return
}

func (s *Orm) assignSingleModel(modelVal reflect.Value, queryVal resultItems) (ret reflect.Value, err error) {
	modelInfo, modelErr := s.modelProvider.GetValueModel(modelVal)
	if modelErr != nil {
		err = modelErr
		log.Errorf("GetValueModel failed, err:%s", err.Error())
		return
	}

	offset := 0
	for _, field := range modelInfo.GetFields() {
		fType := field.GetType()
		dependModel, dependErr := s.modelProvider.GetTypeModel(fType)
		if dependErr != nil {
			err = dependErr
			log.Errorf("GetTypeModel failed, err:%s", err.Error())
			return
		}

		if dependModel != nil {
			itemVal, itemErr := s.queryRelation(modelInfo, field)
			if itemErr != nil {
				log.Errorf("queryRelation failed, err:%s", itemErr.Error())
				continue
			}

			field.SetValue(itemVal)

			offset++
			continue
		}

		qVal := reflect.ValueOf(queryVal[offset]).Elem()
		fVal := fType.Interface()
		fVal, err = helper.AssignValue(qVal, fVal)
		if err != nil {
			log.Errorf("assignValue failed, field name:%s, err:%s", field.GetName(), err.Error())
			return
		}

		err = field.SetValue(fVal)
		if err != nil {
			log.Errorf("UpdateValue failed, err:%s", err.Error())
			return
		}

		offset++
	}

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
		log.Errorf("GetTypeModel failed, err:%s", err.Error())
		return
	}

	queryErr := s.queryBatch(sliceModel, sliceVal, filter)
	if queryErr != nil {
		err = queryErr
		log.Errorf("queryBatch failed, err:%s", err.Error())
		return
	}

	return
}
