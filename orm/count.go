package orm

import (
	"database/sql"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) queryCount(modelInfo model.Model, filter model.Filter) (ret int64, err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	sqlStr, sqlErr := builder.BuildCount(filter)
	if sqlErr != nil {
		err = sqlErr
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
			return
		}

		ret = countVal.Int64
	}

	return
}

// Count count entity
func (s *impl) Count(entityModel model.Model, filter model.Filter) (ret int64, err error) {
	queryVal, queryErr := s.queryCount(entityModel, filter)
	if queryErr != nil {
		err = queryErr
		return
	}
	ret = queryVal

	return
}
