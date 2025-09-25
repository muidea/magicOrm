package database

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/model"
)

type Result interface {
	SQL() string
	Args() []any
}

// Builder orm builder
type Builder interface {
	BuildCreateTable(vModel model.Model) (Result, *cd.Error)
	BuildDropTable(vModel model.Model) (Result, *cd.Error)
	BuildInsert(vModel model.Model) (Result, *cd.Error)
	BuildUpdate(vModel model.Model) (Result, *cd.Error)
	BuildDelete(vModel model.Model) (Result, *cd.Error)
	BuildQuery(vModel model.Model, vFilter model.Filter) (Result, *cd.Error)
	BuildCount(vModel model.Model, vFilter model.Filter) (Result, *cd.Error)

	BuildCreateRelationTable(vModel model.Model, vField model.Field) (Result, *cd.Error)
	BuildDropRelationTable(vModel model.Model, vField model.Field) (Result, *cd.Error)
	BuildInsertRelation(vModel model.Model, vField model.Field, rModel model.Model) (Result, *cd.Error)
	BuildDeleteRelation(vModel model.Model, vField model.Field) (Result, Result, *cd.Error)
	BuildQueryRelation(vModel model.Model, vField model.Field) (Result, *cd.Error)

	BuildModuleValueHolder(vModel model.Model) ([]any, *cd.Error)
}
