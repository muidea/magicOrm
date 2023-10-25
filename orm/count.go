package orm

import (
	"database/sql"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) queryCount(vFilter model.Filter) (ret int64, err error) {
	builderVal := builder.NewBuilder(vFilter.MaskModel(), s.modelProvider, s.specialPrefix)
	sqlStr, sqlErr := builderVal.BuildCount(vFilter)
	if sqlErr != nil {
		err = sqlErr
		log.Errorf("queryCount failed, builderVal.BuildCount error:%s", err.Error())
		return
	}

	err = s.executor.Query(sqlStr)
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

func (s *impl) Count(vFilter model.Filter) (ret int64, re *cd.Result) {
	if vFilter == nil {
		re = cd.NewError(cd.IllegalParam, "illegal filter value")
		return
	}

	queryVal, queryErr := s.queryCount(vFilter)
	if queryErr != nil {
		re = cd.NewError(cd.UnExpected, queryErr.Error())
		log.Errorf("Count failed, s.queryCount error:%s", queryErr.Error())
		return
	}

	ret = queryVal
	return
}
