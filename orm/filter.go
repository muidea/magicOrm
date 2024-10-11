package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) batchQuery(vFilter model.Filter) (ret []model.Model, err *cd.Result) {
	vModel := vFilter.MaskModel()
	hContext := codec.New(s.modelProvider, s.specialPrefix)
	hBuilder := builder.NewBuilder(vModel, hContext)
	queryValueList, queryErr := s.innerQuery(hBuilder, vModel, vFilter)
	if queryErr != nil {
		err = queryErr
		if err.Fail() {
			log.Errorf("batchQuery failed, s.innerQuery error:%s", err.Error())
		} else if err.Warn() {
			log.Warnf("batchQuery failed, s.innerQuery error:%s", err.Error())
		}
		return
	}

	sliceValue := []model.Model{}
	for idx := 0; idx < len(queryValueList); idx++ {
		modelVal, modelErr := s.innerAssign(hBuilder, vModel.Copy(false), queryValueList[idx], 0)
		if modelErr != nil {
			err = modelErr
			if err.Fail() {
				log.Errorf("batchQuery failed, s.innerAssign error:%s", err.Error())
			} else if err.Warn() {
				log.Warnf("batchQuery failed, s.innerAssign error:%s", err.Error())
			}
			return
		}

		sliceValue = append(sliceValue, modelVal)
	}

	ret = sliceValue
	return
}

// BatchQuery batch query
func (s *impl) BatchQuery(vFilter model.Filter) (ret []model.Model, err *cd.Result) {
	if vFilter == nil {
		err = cd.NewError(cd.IllegalParam, "illegal model value")
		return
	}

	queryVal, queryErr := s.batchQuery(vFilter)
	if queryErr != nil {
		err = queryErr
		if err.Fail() {
			log.Errorf("BatchQuery failed, s.batchQuery error:%s", err.Error())
		} else if err.Warn() {
			log.Warnf("BatchQuery failed, s.batchQuery error:%s", err.Error())
		}
		return
	}

	ret = queryVal
	return
}
