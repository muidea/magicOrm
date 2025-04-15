package orm

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
)

// BatchQuery batch query
func (s *impl) BatchQuery(filter model.Filter) (ret []model.Model, err *cd.Error) {
	if filter == nil {
		err = cd.NewError(cd.IllegalParam, "illegal model value")
		log.Errorf("BatchQuery failed, illegal model value")
		return
	}

	vQueryRunner := NewQueryRunner(filter.MaskModel(), s.executor, s.modelProvider, s.modelCodec, true, 0)
	queryVal, queryErr := vQueryRunner.Query(filter)
	if queryErr != nil {
		err = queryErr
		log.Errorf("BatchQuery failed, vQueryRunner.Query error:%v", err.Error())
		return
	}

	ret = queryVal
	return
}
