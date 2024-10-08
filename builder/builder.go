package builder

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database/mysql"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

// Builder orm builder
type Builder interface {
	BuildCreateTable() (string, *cd.Result)
	BuildDropTable() (string, *cd.Result)
	BuildInsert() (string, *cd.Result)
	BuildUpdate() (string, *cd.Result)
	BuildDelete() (string, *cd.Result)
	BuildQuery(filter model.Filter) (string, *cd.Result)
	BuildCount(filter model.Filter) (string, *cd.Result)

	BuildCreateRelationTable(field model.Field, rModel model.Model) (string, *cd.Result)
	BuildDropRelationTable(field model.Field, rModel model.Model) (string, *cd.Result)
	BuildInsertRelation(field model.Field, rModel model.Model) (string, *cd.Result)
	BuildDeleteRelation(field model.Field, rModel model.Model) (string, string, *cd.Result)
	BuildQueryRelation(field model.Field, rModel model.Model) (string, *cd.Result)

	GetFieldPlaceHolder(field model.Field) (interface{}, *cd.Result)
}

// NewBuilder new builder
func NewBuilder(vModel model.Model, modelProvider provider.Provider, prefix string) Builder {
	return mysql.New(vModel, modelProvider, prefix)
}
