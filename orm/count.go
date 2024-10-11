package orm

import (
	"database/sql"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) queryCount(vFilter model.Filter) (ret int64, err *cd.Result) {
	vModel := vFilter.MaskModel()
	hContext := codec.New(s.modelProvider, s.specialPrefix)
	hBuilder := builder.NewBuilder(vModel, hContext)
	countResult, countErr := hBuilder.BuildCount(vFilter)
	if countErr != nil {
		err = countErr
		log.Errorf("queryCount failed, hBuilder.BuildCount error:%s", err.Error())
		return
	}

	_, err = s.executor.Query(countResult.SQL(), false, countResult.Args()...)
	if err != nil {
		return
	}
	defer s.executor.Finish()

	if s.executor.Next() {
		var countVal sql.NullInt64
		err = s.executor.GetField(&countVal)
		if err != nil {
			log.Errorf("queryCount failed, s.executor.GetField error:%s", err.Error())
			return
		}

		ret = countVal.Int64
	}

	return
}

func (s *impl) Count(vFilter model.Filter) (ret int64, err *cd.Result) {
	if vFilter == nil {
		err = cd.NewError(cd.IllegalParam, "illegal filter value")
		return
	}

	queryVal, queryErr := s.queryCount(vFilter)
	if queryErr != nil {
		err = queryErr
		log.Errorf("Count failed, s.queryCount error:%s", err.Error())
		return
	}

	ret = queryVal
	return
}
