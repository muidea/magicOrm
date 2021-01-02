package orm

import (
	"database/sql"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *Orm) queryCount(modelInfo model.Model, filter model.Filter) (ret int64, err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	sqlStr, sqlErr := builder.BuildCount(filter)
	if sqlErr != nil {
		err = sqlErr
		log.Errorf("BuildCount failed, err:%s", err.Error())
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
func (s *Orm) Count(entity interface{}, filter model.Filter) (ret int64, err error) {
	entityModel, entityErr := s.modelProvider.GetEntityModel(entity)
	if entityErr != nil {
		err = entityErr
		log.Errorf("GetEntityModel failed, err:%s", err.Error())
		return
	}

	queryVal, queryErr := s.queryCount(entityModel, filter)
	if queryErr != nil {
		err = queryErr
		log.Errorf("queryCount failed, err:%s", err.Error())
		return
	}
	ret = queryVal

	return
}
