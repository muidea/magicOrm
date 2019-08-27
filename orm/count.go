package orm

import (
	dbsql "database/sql"
	"log"
	"reflect"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

func (s *Orm) queryCount(modelInfo model.Model, filter model.Filter) (ret int64, err error) {
	builder := builder.NewBuilder(modelInfo, s.modelProvider)
	sql, sqlErr := builder.BuildCount(filter)
	if sqlErr != nil {
		err = sqlErr
		log.Printf("BuildCount failed, err:%s", err.Error())
		return
	}

	s.executor.Query(sql)
	defer s.executor.Finish()
	if s.executor.Next() {
		var countVal dbsql.NullInt64
		s.executor.GetField(&countVal)
		ret = countVal.Int64
	}

	return
}

// Count count entity
func (s *Orm) Count(entity interface{}, filter model.Filter) (ret int64, err error) {
	entityVal := reflect.ValueOf(entity)
	modelInfo, modelErr := s.modelProvider.GetValueModel(entityVal)
	if modelErr != nil {
		err = modelErr
		log.Printf("GetValueModel failed, err:%s", err.Error())
		return
	}

	queryVal, queryErr := s.queryCount(modelInfo, filter)
	if queryErr != nil {
		err = queryErr
		log.Printf("queryCount failed, err:%s", err.Error())
		return
	}
	ret = queryVal

	return
}
