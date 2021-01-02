package orm

import (
	"fmt"
	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

type resultItems []interface{}

func (s *Orm) queryBatch(elemModel model.Model, sliceValue model.Value, filter model.Filter) (err error) {
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

	var queryList []resultItems
	func() {
		builder := builder.NewBuilder(elemModel, s.modelProvider)
		sql, sqlErr := builder.BuildBatchQuery(filter)
		if sqlErr != nil {
			err = sqlErr
			log.Errorf("BuildBatchQuery failed, err:%s", err.Error())
			return
		}

		err = s.executor.Query(sql)
		if err != nil {
			return
		}
		defer s.executor.Finish()
		for s.executor.Next() {
			modelItems, modelErr := s.getModelItems(maskModel, builder)
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
	}()
	if err != nil {
		return
	}

	for idx := 0; idx < len(queryList); idx++ {
		modelVal, err := s.assignSingleModel(maskModel.Copy(), queryList[idx])
		if err != nil {
			log.Errorf("assignSingle model failed, err:%s", err.Error())
			return
		}

		sliceValue, err = s.modelProvider.AppendSliceValue(sliceValue, modelVal)
		if err != nil {
			log.Errorf("append slice value failed, err:%s", err.Error())
			return
		}
	}

	return
}

func (s *Orm) assignSingleModel(modelVal model.Model, queryVal resultItems) (ret model.Value, err error) {
	offset := 0
	for _, field := range modelVal.GetFields() {
		if field.GetValue().IsNil() {
			continue
		}

		fType := field.GetType()
		if !fType.IsBasic() {
			itemVal, itemErr := s.queryRelation(modelVal, field)
			if itemErr != nil {
				log.Errorf("queryRelation failed, err:%s", itemErr.Error())
				continue
			}

			field.UpdateValue(itemVal)

			//offset++
			continue
		}

		fVal := fType.Interface(s.stripSlashes(fType, queryVal[offset]))
		err = field.UpdateValue(fVal)
		if err != nil {
			log.Errorf("UpdateValue failed, err:%s", err.Error())
			return
		}

		offset++
	}

	ret = modelVal.Interface()
	return
}

// BatchQuery batch query
func (s *Orm) BatchQuery(sliceEntity interface{}, filter model.Filter) (err error) {
	entityModel, entityErr := s.modelProvider.GetEntityModel(sliceEntity)
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

	queryErr := s.queryBatch(entityModel, sliceEntityVal, filter)
	if queryErr != nil {
		err = queryErr
		log.Errorf("queryBatch failed, err:%s", err.Error())
		return
	}

	return
}
