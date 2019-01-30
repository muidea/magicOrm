package builder

import (
	"muidea.com/magicOrm/database/mysql"
	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/provider"
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
	BuildBatchQuery(filter model.Filter) (string, error)

	GetRelationTableName(fieldName string, relationInfo model.Model) string
	BuildCreateRelationSchema(fieldName string, relationInfo model.Model) (string, error)
	BuildDropRelationSchema(fieldName string, relationInfo model.Model) (string, error)
	BuildInsertRelation(fieldName string, relationInfo model.Model) (string, error)
	BuildDeleteRelation(fieldName string, relationInfo model.Model) (string, string, error)
	BuildQueryRelation(fieldName string, relationInfo model.Model) (string, error)
}

// NewBuilder new builder
func NewBuilder(modelInfo model.Model, modelProvider provider.Provider) Builder {
	return mysql.New(modelInfo, modelProvider)
}

// EquleOpr EquleOpr
func EquleOpr(name string, val string) string {
	return mysql.EquleOpr(name, val)
}

// NotEquleOpr NotEquleOpr
func NotEquleOpr(name string, val string) string {
	return mysql.NotEquleOpr(name, val)
}

// BelowOpr BelowOpr
func BelowOpr(name string, val string) string {
	return mysql.BelowOpr(name, val)
}

// AboveOpr AboveOpr
func AboveOpr(name string, val string) string {
	return mysql.AboveOpr(name, val)
}

// InOpr InOpr
func InOpr(name string, val string) string {
	return mysql.InOpr(name, val)
}

// NotInOpr NotInOpr
func NotInOpr(name string, val string) string {
	return mysql.NotInOpr(name, val)
}

// LikeOpr LikeOpr
func LikeOpr(name string, val string) string {
	return mysql.LikeOpr(name, val)
}
