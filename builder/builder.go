package builder

import (
	"github.com/muidea/magicOrm/database/mysql"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

// Builder orm builder
type Builder interface {
	BuildCreateTable() (string, error)
	BuildDropTable() (string, error)
	BuildInsert() (string, error)
	BuildUpdate() (string, error)
	BuildDelete() (string, error)
	BuildQuery(filter model.Filter) (string, error)
	BuildCount(filter model.Filter) (string, error)

	BuildCreateRelationTable(relationTableName string) (string, error)
	BuildDropRelationTable(relationTableName string) (string, error)
	BuildInsertRelation(field model.Field, rModel model.Model) (string, error)
	BuildDeleteRelation(field model.Field, rModel model.Model) (string, string, error)
	BuildQueryRelation(field model.Field, rModel model.Model) (string, error)

	GetTableName() string
	GetHostTableName(vModel model.Model) string
	GetRelationTableName(field model.Field, rModel model.Model) string
	GetFieldInitializeValue(field model.Field) (interface{}, error)
}

// NewBuilder new builder
func NewBuilder(vModel model.Model, modelProvider provider.Provider, prefix string) Builder {
	return mysql.New(vModel, modelProvider, prefix)
}
