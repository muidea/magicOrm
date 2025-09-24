package builder

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/database/postgres"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

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

	GetModuleValueHolder(vModel model.Model) ([]any, *cd.Error)
	GetFieldValueHolder(vField model.Field) (any, *cd.Error)
	GetRelationFieldValueHolder(vModel model.Model, vField model.Field) (any, *cd.Error)
}

type builderImpl struct {
	builder *postgres.Builder
}

func (s *builderImpl) BuildCreateTable(vModel model.Model) (Result, *cd.Error) {
	return s.builder.BuildCreateTable(vModel)
}

func (s *builderImpl) BuildDropTable(vModel model.Model) (Result, *cd.Error) {
	return s.builder.BuildDropTable(vModel)
}

func (s *builderImpl) BuildInsert(vModel model.Model) (Result, *cd.Error) {
	return s.builder.BuildInsert(vModel)
}

func (s *builderImpl) BuildUpdate(vModel model.Model) (Result, *cd.Error) {
	return s.builder.BuildUpdate(vModel)
}

func (s *builderImpl) BuildDelete(vModel model.Model) (Result, *cd.Error) {
	return s.builder.BuildDelete(vModel)
}

func (s *builderImpl) BuildQuery(vModel model.Model, vFilter model.Filter) (Result, *cd.Error) {
	return s.builder.BuildQuery(vModel, vFilter)
}

func (s *builderImpl) GetModuleValueHolder(vModel model.Model) ([]any, *cd.Error) {
	return s.builder.GetModuleValueHolder(vModel)
}

func (s *builderImpl) GetFieldValueHolder(vField model.Field) (any, *cd.Error) {
	return s.builder.GetFieldPlaceHolder(vField)
}

func (s *builderImpl) BuildCount(vModel model.Model, vFilter model.Filter) (Result, *cd.Error) {
	return s.builder.BuildCount(vModel, vFilter)
}

func (s *builderImpl) BuildCreateRelationTable(vModel model.Model, vField model.Field) (Result, *cd.Error) {
	return s.builder.BuildCreateRelationTable(vModel, vField)
}

func (s *builderImpl) BuildDropRelationTable(vModel model.Model, vField model.Field) (Result, *cd.Error) {
	return s.builder.BuildDropRelationTable(vModel, vField)
}

func (s *builderImpl) BuildInsertRelation(vModel model.Model, vField model.Field, rModel model.Model) (Result, *cd.Error) {
	return s.builder.BuildInsertRelation(vModel, vField, rModel)
}

func (s *builderImpl) BuildDeleteRelation(vModel model.Model, vField model.Field) (Result, Result, *cd.Error) {
	return s.builder.BuildDeleteRelation(vModel, vField)
}

func (s *builderImpl) BuildQueryRelation(vModel model.Model, vField model.Field) (Result, *cd.Error) {
	return s.builder.BuildQueryRelation(vModel, vField)
}

func (s *builderImpl) GetRelationFieldValueHolder(vModel model.Model, vField model.Field) (any, *cd.Error) {
	return s.builder.GetRelationFieldValueHolder(vModel, vField)
}

// NewBuilder new builder
func NewBuilder(provider provider.Provider, codec codec.Codec) Builder {
	return &builderImpl{
		builder: postgres.New(provider, codec),
	}
}
