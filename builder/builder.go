package builder

import (
	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/mysql"
)

// Builder orm builder
type Builder interface {
	BuildCreateSchema() (string, error)
	BuildDropSchema() (string, error)
	BuildInsert() (string, error)
	BuildUpdate() (string, error)
	BuildDelete() (string, error)
	BuildQuery() (string, error)
	GetTableName() string

	GetRelationTableName(fieldName string, relationInfo model.StructInfo) string
	BuildCreateRelationSchema(fieldName string, relationInfo model.StructInfo) (string, error)
	BuildDropRelationSchema(fieldName string, relationInfo model.StructInfo) (string, error)
	BuildInsertRelation(fieldName string, relationInfo model.StructInfo) (string, error)
	BuildDeleteRelation(fieldName string, relationInfo model.StructInfo) (string, string, error)
	BuildQueryRelation(fieldName string, relationInfo model.StructInfo) (string, error)
}

// NewBuilder new builder
func NewBuilder(structInfo model.StructInfo) Builder {
	return mysql.New(structInfo)
}
