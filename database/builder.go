package database

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
)

type Result interface {
	SQL() string
	Args() []any
}

// Builder orm builder
type Builder interface {
	BuildCreateTable(vModel models.Model) (Result, *cd.Error)
	BuildDropTable(vModel models.Model) (Result, *cd.Error)
	BuildInsert(vModel models.Model) (Result, *cd.Error)
	BuildUpdate(vModel models.Model) (Result, *cd.Error)
	BuildDelete(vModel models.Model) (Result, *cd.Error)
	BuildQuery(vModel models.Model, vFilter models.Filter) (Result, *cd.Error)
	BuildCount(vModel models.Model, vFilter models.Filter) (Result, *cd.Error)

	BuildCreateRelationTable(vModel models.Model, vField models.Field) (Result, *cd.Error)
	BuildDropRelationTable(vModel models.Model, vField models.Field) (Result, *cd.Error)
	BuildInsertRelation(vModel models.Model, vField models.Field, rModel models.Model) (Result, *cd.Error)
	BuildDeleteRelation(vModel models.Model, vField models.Field) (Result, Result, *cd.Error)
	BuildQueryRelation(vModel models.Model, vField models.Field) (Result, *cd.Error)

	BuildModuleValueHolder(vModel models.Model) ([]any, *cd.Error)
}
