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

type resultItems []interface{}

func (s *Orm) queryBatch(elemType model.Type, elemModel model.Model, sliceValue reflect.Value, filter model.Filter) (err error) {
	var maskModel model.Model
	if filter != nil {
		maskModel, err = filter.MaskModel()
		if err != nil {
			log.Errorf("Get MaskModel failed, err:%s", err.Error())
			return
		}

		if maskModel != nil {
			if maskModel.GetName() != elemModel.GetName() || maskModel.GetPkgPath() != elemModel.GetPkgPath() {
				err = fmt.Errorf("illegal value mask")
				return
			}
		}
	}
	if maskModel == nil {
		maskModel = elemModel
	}

	builder := builder.NewBuilder(elemModel, s.modelProvider)
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

	for idx := 0; idx < len(queryList); idx++ {
		modelVal := maskModel.Interface()
		modelVal, err = s.assignSingleModel(modelVal, queryList[idx])
		if err != nil {
			log.Errorf("assignSingle model failed, err:%s", err.Error())
			return
		}

		if elemType.IsPtrType() {
			modelVal = modelVal.Addr()
		}

		sliceValue, err = s.modelProvider.AppendSliceValue(sliceValue, modelVal)
		if err != nil {
			log.Errorf("append slice value failed, err:%s", err.Error())
			return
		}
	}

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

			field.UpdateValue(itemVal)

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

		err = field.UpdateValue(fVal)
		if err != nil {
			log.Errorf("UpdateValue failed, err:%s", err.Error())
			return
		}

		offset++
	}

	ret = modelVal
	return
}

// BatchQuery batch query
func (s *Orm) BatchQuery(sliceEntity interface{}, filter model.Filter) (err error) {
	sliceEntityVal := reflect.ValueOf(sliceEntity)
	sliceEntityVal = reflect.Indirect(sliceEntityVal)
	entityType, entityErr := s.modelProvider.GetEntityType(sliceEntity)
	if entityErr != nil {
		err = entityErr
		return
	}
	if !util.IsSliceType(entityType.GetValue()) || !entityType.IsPtrType() {
		err = fmt.Errorf("illegal entity value")
		return
	}

	elemType := entityType.Elem()
	elemModel, elemErr := s.modelProvider.GetTypeModel(elemType)
	if elemErr != nil {
		err = elemErr
		log.Errorf("GetTypeModel failed, err:%s", err.Error())
		return
	}

	queryErr := s.queryBatch(elemType, elemModel, sliceEntityVal, filter)
	if queryErr != nil {
		err = queryErr
		log.Errorf("queryBatch failed, err:%s", err.Error())
		return
	}

	return
}
