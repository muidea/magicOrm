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
	// BuildAddColumn emits the safe, additive part of a model evolution. The
	// caller is responsible for rejecting destructive changes before executing
	// the result.
	BuildAddColumn(vModel models.Model, vField models.Field) (Result, *cd.Error)
	BuildCreateUniqueConstraint(vModel models.Model, constraint models.UniqueConstraint) (Result, *cd.Error)
	BuildCreateIndex(vModel models.Model, index models.Index) (Result, *cd.Error)
	BuildInsert(vModel models.Model) (Result, *cd.Error)
	BuildUpdate(vModel models.Model) (Result, *cd.Error)
	// BuildUpdateWithFilter builds a single-table conditional update. The
	// filter is intentionally restricted to basic fields so callers cannot
	// turn a state transition into a relation-driven multi-statement write.
	BuildUpdateWithFilter(vModel models.Model, filter models.Filter) (Result, *cd.Error)
	BuildDelete(vModel models.Model) (Result, *cd.Error)
	BuildQuery(vModel models.Model, vFilter models.Filter) (Result, *cd.Error)
	BuildCount(vModel models.Model, vFilter models.Filter) (Result, *cd.Error)

	BuildCreateRelationTable(vModel models.Model, vField models.Field) (Result, *cd.Error)
	BuildDropRelationTable(vModel models.Model, vField models.Field) (Result, *cd.Error)
	BuildInsertRelation(vModel models.Model, vField models.Field, rModel models.Model) (Result, *cd.Error)
	BuildDeleteRelation(vModel models.Model, vField models.Field) (Result, Result, *cd.Error)
	BuildDeleteRelationByRights(vModel models.Model, vField models.Field, rightIDs []any) (Result, *cd.Error)
	BuildQueryRelation(vModel models.Model, vField models.Field) (Result, *cd.Error)
	BuildBatchQueryRelation(vModel models.Model, vField models.Field, leftIDs []any) (Result, *cd.Error)

	BuildModuleValueHolder(vModel models.Model) ([]any, *cd.Error)
}
