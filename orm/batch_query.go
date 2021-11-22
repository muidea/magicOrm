package orm

import (
	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) queryBatch(elemModel model.Model, filter model.Filter) (ret []model.Model, err error) {
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

	sliceValue := []model.Model{}
	for idx := 0; idx < len(queryValueList); idx++ {
		modelVal, modelErr := s.assignSingleModel(maskModel.Copy(), queryValueList[idx], 0)
		if modelErr != nil {
			err = modelErr
			return
		}

		sliceValue = append(sliceValue, modelVal)
	}

	ret = sliceValue
	return
}

// BatchQuery batch query
func (s *impl) BatchQuery(entityModel model.Model, filter model.Filter) (ret []model.Model, err error) {
	queryVal, queryErr := s.queryBatch(entityModel, filter)
	if queryErr != nil {
		err = queryErr
		return
	}

	ret = queryVal
	return
}
