package builder

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database/context"
	"github.com/muidea/magicOrm/database/mysql"
	"github.com/muidea/magicOrm/model"
)

// Builder orm builder
type Builder interface {
	BuildCreateTable() (context.BuildResult, *cd.Result)
	BuildDropTable() (context.BuildResult, *cd.Result)
	BuildInsert() (context.BuildResult, *cd.Result)
	BuildUpdate() (context.BuildResult, *cd.Result)
	BuildDelete() (context.BuildResult, *cd.Result)
	BuildQuery(filter model.Filter) (context.BuildResult, *cd.Result)
	BuildCount(filter model.Filter) (context.BuildResult, *cd.Result)

	BuildCreateRelationTable(field model.Field, rModel model.Model) (context.BuildResult, *cd.Result)
	BuildDropRelationTable(field model.Field, rModel model.Model) (context.BuildResult, *cd.Result)
	BuildInsertRelation(field model.Field, rModel model.Model) (context.BuildResult, *cd.Result)
	BuildDeleteRelation(field model.Field, rModel model.Model) (context.BuildResult, context.BuildResult, *cd.Result)
	BuildQueryRelation(field model.Field, rModel model.Model) (context.BuildResult, *cd.Result)

	GetFieldPlaceHolder(field model.Field) (any, *cd.Result)
}

// NewBuilder new builder
func NewBuilder(vModel model.Model, context context.Context) Builder {
	return mysql.New(vModel, context)
}
