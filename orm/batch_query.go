package orm

import (
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

	queryValueList, queryErr := s.innerQuery(maskModel, filter)
	if queryErr != nil {
		err = queryErr
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
