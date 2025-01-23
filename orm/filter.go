package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
)

// BatchQuery batch query
func (s *impl) BatchQuery(filter model.Filter) (ret []model.Model, err *cd.Result) {
	if filter == nil {
		err = cd.NewResult(cd.IllegalParam, "illegal model value")
		return
	}

	vModel := filter.MaskModel()
	vQueryRunner := NewQueryRunner(vModel, s.executor, s.modelProvider, s.modelCodec, true, 0)
	queryVal, queryErr := vQueryRunner.Query(filter)
	if queryErr != nil {
		err = queryErr
		if err.Fail() {
			log.Errorf("BatchQuery failed, vQueryRunner.Query error:%v", err.Error())
		} else if err.Warn() {
			log.Warnf("BatchQuery failed, vQueryRunner.Query error:%v", err.Error())
		}
		return
	}

	ret = queryVal
	return
}
