package builder

import (
	"muidea.com/magicOrm/filter"
	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/mysql"
)

// Builder orm builder
type Builder interface {
	GetTableName() string
	BuildCreateSchema() (string, error)
	BuildDropSchema() (string, error)
	BuildInsert() (string, error)
	BuildUpdate() (string, error)
	BuildDelete() (string, error)
	BuildQuery() (string, error)
	BuildBatchQuery(filter filter.Filter) (string, error)

	GetRelationTableName(fieldName string, relationInfo model.Model) string
	BuildCreateRelationSchema(fieldName string, relationInfo model.Model) (string, error)
	BuildDropRelationSchema(fieldName string, relationInfo model.Model) (string, error)
	BuildInsertRelation(fieldName string, relationInfo model.Model) (string, error)
	BuildDeleteRelation(fieldName string, relationInfo model.Model) (string, string, error)
	BuildQueryRelation(fieldName string, relationInfo model.Model) (string, error)
}

// NewBuilder new builder
func NewBuilder(structInfo model.Model) Builder {
	return mysql.New(structInfo)
}
