package orm

import (
	"github.com/muidea/magicOrm/model"
)

func (s *impl) queryBatch(vFilter model.Filter) (ret []model.Model, err error) {
	vModel := vFilter.MaskModel()
	queryValueList, queryErr := s.innerQuery(vModel, vFilter)
	if queryErr != nil {
		err = queryErr
		return
	}

	sliceValue := []model.Model{}
	for idx := 0; idx < len(queryValueList); idx++ {
		modelVal, modelErr := s.assignSingleModel(vModel.Copy(), queryValueList[idx], 0)
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
func (s *impl) BatchQuery(filter model.Filter) (ret []model.Model, err error) {
	queryVal, queryErr := s.queryBatch(filter)
	if queryErr != nil {
		err = queryErr
		return
	}

	ret = queryVal
	return
}
