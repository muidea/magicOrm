package orm

import (
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) queryBatch(vFilter model.Filter) (ret []model.Model, err error) {
	vModel := vFilter.MaskModel()
	queryValueList, queryErr := s.innerQuery(vModel, vFilter)
	if queryErr != nil {
		err = queryErr
		log.Errorf("queryBatch failed, s.innerQuery error:%s", err.Error())
		return
	}

	sliceValue := []model.Model{}
	for idx := 0; idx < len(queryValueList); idx++ {
		modelVal, modelErr := s.innerAssign(vModel.Copy(), queryValueList[idx], 0)
		if modelErr != nil {
			err = modelErr
			log.Errorf("queryBatch failed, s.innerAssign error:%s", err.Error())
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
		log.Errorf("BatchQuery failed, s.queryBatch error:%s", err.Error())
		return
	}

	ret = queryVal
	return
}
