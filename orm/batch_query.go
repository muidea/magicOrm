package orm

import (
	"fmt"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

func (s *Orm) queryBatch(elemModel model.Model, sliceValue model.Value, filter model.Filter) (ret model.Value, err error) {
	var maskModel model.Model
	if filter != nil {
		maskModel = filter.MaskModel()
	}
	if maskModel == nil {
		maskModel = elemModel
	}

	var queryValueList []resultItems
	func() {
		builder := builder.NewBuilder(elemModel, s.modelProvider)
		sqlStr, sqlErr := builder.BuildBatchQuery(filter)
		if sqlErr != nil {
			err = sqlErr
			log.Errorf("BuildBatchQuery failed, err:%s", err.Error())
			return
		}

		err = s.executor.Query(sqlStr)
		if err != nil {
			return
		}
		defer s.executor.Finish()
		for s.executor.Next() {
			itemValues, itemErr := s.getInitializeValue(maskModel, builder)
			if itemErr != nil {
				err = itemErr
				return
			}

			err = s.executor.GetField(itemValues...)
			if err != nil {
				return
			}

			queryValueList = append(queryValueList, itemValues)
		}
	}()
	if err != nil {
		return
	}

	for idx := 0; idx < len(queryValueList); idx++ {
		modelVal, modelErr := s.assignSingleModel(maskModel.Copy(), queryValueList[idx])
		if modelErr != nil {
			err = modelErr
			log.Errorf("assignSingle model failed, err:%s", err.Error())
			return
		}

		sliceValue, err = s.modelProvider.AppendSliceValue(sliceValue, modelVal)
		if err != nil {
			log.Errorf("append slice value failed, err:%s", err.Error())
			return
		}
	}

	ret = sliceValue
	return
}

// BatchQuery batch query
func (s *Orm) BatchQuery(sliceEntity interface{}, filter model.Filter) (err error) {
	entityType, entityErr := s.modelProvider.GetEntityType(sliceEntity)
	if entityErr != nil {
		err = entityErr
		log.Errorf("GetEntityType failed, err:%s", err.Error())
		return
	}

	if !util.IsSliceType(entityType.GetValue()) {
		err = fmt.Errorf("illegal entity, must be a slice ptr")
		return
	}

	entityModel, entityErr := s.modelProvider.GetTypeModel(entityType.Elem())
	if entityErr != nil {
		err = entityErr
		log.Errorf("GetEntityModel failed, err:%s", err.Error())
		return
	}

	sliceEntityVal, sliceEntityErr := s.modelProvider.GetEntityValue(sliceEntity)
	if sliceEntityErr != nil {
		err = entityErr
		log.Errorf("GetEntityValue failed, err:%s", err.Error())
		return
	}

	queryVal, queryErr := s.queryBatch(entityModel, sliceEntityVal, filter)
	if queryErr != nil {
		err = queryErr
		log.Errorf("queryBatch failed, err:%s", err.Error())
		return
	}

	sliceEntityVal.Set(queryVal.Get())
	return
}
