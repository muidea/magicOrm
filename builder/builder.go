package builder

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/database/mysql"
	"github.com/muidea/magicOrm/model"
)

// Builder orm builder
type Builder interface {
	BuildCreateTable() (codec.BuildResult, *cd.Result)
	BuildDropTable() (codec.BuildResult, *cd.Result)
	BuildInsert() (codec.BuildResult, *cd.Result)
	BuildUpdate() (codec.BuildResult, *cd.Result)
	BuildDelete() (codec.BuildResult, *cd.Result)
	BuildQuery(filter model.Filter) (codec.BuildResult, *cd.Result)
	BuildCount(filter model.Filter) (codec.BuildResult, *cd.Result)

	BuildCreateRelationTable(field model.Field, rModel model.Model) (codec.BuildResult, *cd.Result)
	BuildDropRelationTable(field model.Field, rModel model.Model) (codec.BuildResult, *cd.Result)
	BuildInsertRelation(field model.Field, rModel model.Model) (codec.BuildResult, *cd.Result)
	BuildDeleteRelation(field model.Field, rModel model.Model) (codec.BuildResult, codec.BuildResult, *cd.Result)
	BuildQueryRelation(field model.Field, rModel model.Model) (codec.BuildResult, *cd.Result)

	GetFieldPlaceHolder(field model.Field) (any, *cd.Result)
}

// NewBuilder new builder
func NewBuilder(vModel model.Model, context codec.Codec) Builder {
	return mysql.New(vModel, context)
}
