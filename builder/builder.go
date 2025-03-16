package builder

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/database/mysql"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

// Builder orm builder
type Builder interface {
	BuildCreateTable(vModel model.Model) (Result, *cd.Result)
	BuildDropTable(vModel model.Model) (Result, *cd.Result)
	BuildInsert(vModel model.Model) (Result, *cd.Result)
	BuildUpdate(vModel model.Model) (Result, *cd.Result)
	BuildDelete(vModel model.Model) (Result, *cd.Result)
	BuildQuery(vModel model.Model, vFilter model.Filter) (Result, *cd.Result)
	BuildQueryPlaceHolder(vModel model.Model) ([]any, *cd.Result)
	BuildCount(vModel model.Model, vFilter model.Filter) (Result, *cd.Result)

	BuildCreateRelationTable(vModel model.Model, vField model.Field) (Result, *cd.Result)
	BuildDropRelationTable(vModel model.Model, vField model.Field) (Result, *cd.Result)
	BuildInsertRelation(vModel model.Model, vField model.Field, rModel model.Model) (Result, *cd.Result)
	BuildDeleteRelation(vModel model.Model, vField model.Field) (Result, Result, *cd.Result)
	BuildQueryRelation(vModel model.Model, vField model.Field) (Result, *cd.Result)
	BuildQueryRelationPlaceHolder(vModel model.Model, vField model.Field) (any, *cd.Result)
}

type builderImpl struct {
	builder *mysql.Builder
}

func (s *builderImpl) BuildCreateTable(vModel model.Model) (Result, *cd.Result) {
	return s.builder.BuildCreateTable(vModel)
}

func (s *builderImpl) BuildDropTable(vModel model.Model) (Result, *cd.Result) {
	return s.builder.BuildDropTable(vModel)
}

func (s *builderImpl) BuildInsert(vModel model.Model) (Result, *cd.Result) {
	return s.builder.BuildInsert(vModel)
}

func (s *builderImpl) BuildUpdate(vModel model.Model) (Result, *cd.Result) {
	return s.builder.BuildUpdate(vModel)
}

func (s *builderImpl) BuildDelete(vModel model.Model) (Result, *cd.Result) {
	return s.builder.BuildDelete(vModel)
}

func (s *builderImpl) BuildQuery(vModel model.Model, vFilter model.Filter) (Result, *cd.Result) {
	return s.builder.BuildQuery(vModel, vFilter)
}

func (s *builderImpl) BuildQueryPlaceHolder(vModel model.Model) ([]any, *cd.Result) {
	return s.builder.BuildQueryPlaceHolder(vModel)
}

func (s *builderImpl) BuildCount(vModel model.Model, vFilter model.Filter) (Result, *cd.Result) {
	return s.builder.BuildCount(vModel, vFilter)
}

func (s *builderImpl) BuildCreateRelationTable(vModel model.Model, vField model.Field) (Result, *cd.Result) {
	return s.builder.BuildCreateRelationTable(vModel, vField)
}

func (s *builderImpl) BuildDropRelationTable(vModel model.Model, vField model.Field) (Result, *cd.Result) {
	return s.builder.BuildDropRelationTable(vModel, vField)
}

func (s *builderImpl) BuildInsertRelation(vModel model.Model, vField model.Field, rModel model.Model) (Result, *cd.Result) {
	return s.builder.BuildInsertRelation(vModel, vField,rModel)
}

func (s *builderImpl) BuildDeleteRelation(vModel model.Model, vField model.Field) (Result, Result, *cd.Result) {
	return s.builder.BuildDeleteRelation(vModel, vField)
}

func (s *builderImpl) BuildQueryRelation(vModel model.Model, vField model.Field) (Result, *cd.Result) {
	return s.builder.BuildQueryRelation(vModel, vField)
}

func (s *builderImpl) BuildQueryRelationPlaceHolder(vModel model.Model, vField model.Field) (any, *cd.Result) {
	return s.builder.BuildQueryRelationPlaceHolder(vModel, vField)
}

// NewBuilder new builder
func NewBuilder(provider provider.Provider, codec codec.Codec) Builder {
	return &builderImpl{
		builder: mysql.New(provider, codec),
	}
}
