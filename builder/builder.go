package builder

import (
	"github.com/muidea/magicOrm/database/mysql"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

// Builder orm builder
type Builder interface {
	GetTableName() string
	BuildCreateSchema() (string, error)
	BuildDropSchema() (string, error)
	BuildInsert() (string, error)
	BuildUpdate() (string, error)
	BuildDelete() (string, error)
	BuildQuery(filter model.Filter) (string, error)
	BuildCount(filter model.Filter) (string, error)
	BuildBatchQuery(filter model.Filter) (string, error)

	GetRelationTableName(fieldName string, relationInfo model.Model) string
	BuildCreateRelationSchema(fieldName string, relationInfo model.Model) (string, error)
	BuildDropRelationSchema(fieldName string, relationInfo model.Model) (string, error)
	BuildInsertRelation(fieldName string, relationInfo model.Model) (string, error)
	BuildDeleteRelation(fieldName string, relationInfo model.Model) (string, string, error)
	BuildQueryRelation(fieldName string, relationInfo model.Model) (string, error)

	GetInitializeValue(field model.Field) (interface{}, error)
}

// NewBuilder new builder
func NewBuilder(modelInfo model.Model, modelProvider provider.Provider) Builder {
	return mysql.New(modelInfo, modelProvider)
}

// EqualOpr EqualOpr
func EqualOpr(name string, val interface{}) string {
	return mysql.EqualOpr(name, val)
}

// NotEqualOpr NotEqualOpr
func NotEqualOpr(name string, val interface{}) string {
	return mysql.NotEqualOpr(name, val)
}

// BelowOpr BelowOpr
func BelowOpr(name string, val interface{}) string {
	return mysql.BelowOpr(name, val)
}

// AboveOpr AboveOpr
func AboveOpr(name string, val interface{}) string {
	return mysql.AboveOpr(name, val)
}

// InOpr InOpr
func InOpr(name string, val interface{}) string {
	return mysql.InOpr(name, val)
}

// NotInOpr NotInOpr
func NotInOpr(name string, val interface{}) string {
	return mysql.NotInOpr(name, val)
}

// LikeOpr LikeOpr
func LikeOpr(name string, val interface{}) string {
	return mysql.LikeOpr(name, val)
}
