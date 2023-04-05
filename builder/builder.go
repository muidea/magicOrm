package builder

import (
	"github.com/muidea/magicOrm/database/mysql"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

// Builder orm builder
type Builder interface {
	BuildCreateSchema() (string, error)
	BuildDropSchema() (string, error)
	BuildInsert() (string, error)
	BuildUpdate() (string, error)
	BuildDelete() (string, error)
	BuildQuery(filter model.Filter) (string, error)
	BuildCount(filter model.Filter) (string, error)

	BuildCreateRelationSchema(relationSchema string) (string, error)
	BuildDropRelationSchema(relationSchema string) (string, error)
	BuildInsertRelation(field model.Field, relationInfo model.Model) (string, error)
	BuildDeleteRelation(field model.Field, relationInfo model.Model) (string, string, error)
	BuildQueryRelation(field model.Field, relationInfo model.Model) (string, error)

	GetTableName() string
	GetHostTableName(vModel model.Model) string
	GetRelationTableName(field model.Field, relationInfo model.Model) string
	GetInitializeValue(field model.Field) (interface{}, error)
}

// NewBuilder new builder
func NewBuilder(modelInfo model.Model, modelProvider provider.Provider) Builder {
	return mysql.New(modelInfo, modelProvider)
}

// EqualOpr Equal Opr =
func EqualOpr(name string, val interface{}) string {
	return mysql.EqualOpr(name, val)
}

// NotEqualOpr NotEqual Opr !=
func NotEqualOpr(name string, val interface{}) string {
	return mysql.NotEqualOpr(name, val)
}

// BelowOpr Below Opr <
func BelowOpr(name string, val interface{}) string {
	return mysql.BelowOpr(name, val)
}

// AboveOpr Above Opr >
func AboveOpr(name string, val interface{}) string {
	return mysql.AboveOpr(name, val)
}

// InOpr In Opr in
func InOpr(name string, val interface{}) string {
	return mysql.InOpr(name, val)
}

// NotInOpr NotIn Opr not in
func NotInOpr(name string, val interface{}) string {
	return mysql.NotInOpr(name, val)
}

// LikeOpr Like Opr like
func LikeOpr(name string, val interface{}) string {
	return mysql.LikeOpr(name, val)
}
