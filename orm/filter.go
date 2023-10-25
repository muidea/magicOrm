package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
)

func (s *impl) queryBatch(vFilter model.Filter) (ret []model.Model, re *cd.Result) {
	vModel := vFilter.MaskModel()
	queryValueList, queryErr := s.innerQuery(vModel, vFilter)
	if queryErr != nil {
		re = queryErr
		if queryErr.Fail() {
			log.Errorf("queryBatch failed, s.innerQuery error:%s", queryErr.Error())
		} else if queryErr.Warn() {
			log.Warnf("queryBatch failed, s.innerQuery error:%s", queryErr.Error())
		}
		return
	}

	sliceValue := []model.Model{}
	for idx := 0; idx < len(queryValueList); idx++ {
		modelVal, modelErr := s.innerAssign(vModel.Copy(), queryValueList[idx], 0)
		if modelErr != nil {
			re = modelErr
			if modelErr.Fail() {
				log.Errorf("queryBatch failed, s.innerAssign error:%s", modelErr.Error())
			} else if modelErr.Warn() {
				log.Warnf("queryBatch failed, s.innerAssign error:%s", modelErr.Error())
			}
			return
		}

		sliceValue = append(sliceValue, modelVal)
	}

	ret = sliceValue
	return
}

// BatchQuery batch query
func (s *impl) BatchQuery(vFilter model.Filter) (ret []model.Model, re *cd.Result) {
	if vFilter == nil {
		re = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}

	queryVal, queryErr := s.queryBatch(vFilter)
	if queryErr != nil {
		re = queryErr
		if queryErr.Fail() {
			log.Errorf("BatchQuery failed, s.queryBatch error:%s", queryErr.Error())
		} else if queryErr.Warn() {
			log.Warnf("BatchQuery failed, s.queryBatch error:%s", queryErr.Error())
		}
		return
	}

	ret = queryVal
	return
}
