package builder

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/database/mysql"
	"github.com/muidea/magicOrm/model"
)

// Builder orm builder
type Builder interface {
	BuildCreateTable() (Result, *cd.Result)
	BuildDropTable() (Result, *cd.Result)
	BuildInsert() (Result, *cd.Result)
	BuildUpdate() (Result, *cd.Result)
	BuildDelete() (Result, *cd.Result)
	BuildQuery(filter model.Filter) (Result, *cd.Result)
	BuildCount(filter model.Filter) (Result, *cd.Result)

	BuildCreateRelationTable(field model.Field, rModel model.Model) (Result, *cd.Result)
	BuildDropRelationTable(field model.Field, rModel model.Model) (Result, *cd.Result)
	BuildInsertRelation(field model.Field, rModel model.Model) (Result, *cd.Result)
	BuildDeleteRelation(field model.Field, rModel model.Model) (Result, Result, *cd.Result)
	BuildQueryRelation(field model.Field, rModel model.Model) (Result, *cd.Result)

	GetFieldPlaceHolder(field model.Field) (any, *cd.Result)
}

type builderImpl struct {
	builder *mysql.Builder
}

func (s *builderImpl) BuildCreateTable() (Result, *cd.Result) {
	return s.builder.BuildCreateTable()
}

func (s *builderImpl) BuildDropTable() (Result, *cd.Result) {
	return s.builder.BuildDropTable()
}

func (s *builderImpl) BuildInsert() (Result, *cd.Result) {
	return s.builder.BuildInsert()
}

func (s *builderImpl) BuildUpdate() (Result, *cd.Result) {
	return s.builder.BuildUpdate()
}

func (s *builderImpl) BuildDelete() (Result, *cd.Result) {
	return s.builder.BuildDelete()
}

func (s *builderImpl) BuildQuery(filter model.Filter) (Result, *cd.Result) {
	return s.builder.BuildQuery(filter)
}

func (s *builderImpl) BuildCount(filter model.Filter) (Result, *cd.Result) {
	return s.builder.BuildCount(filter)
}

func (s *builderImpl) BuildCreateRelationTable(field model.Field, rModel model.Model) (Result, *cd.Result) {
	return s.builder.BuildCreateRelationTable(field, rModel)
}

func (s *builderImpl) BuildDropRelationTable(field model.Field, rModel model.Model) (Result, *cd.Result) {
	return s.builder.BuildDropRelationTable(field, rModel)
}

func (s *builderImpl) BuildInsertRelation(field model.Field, rModel model.Model) (Result, *cd.Result) {
	return s.builder.BuildInsertRelation(field, rModel)
}

func (s *builderImpl) BuildDeleteRelation(field model.Field, rModel model.Model) (Result, Result, *cd.Result) {
	return s.builder.BuildDeleteRelation(field, rModel)
}

func (s *builderImpl) BuildQueryRelation(field model.Field, rModel model.Model) (Result, *cd.Result) {
	return s.builder.BuildQueryRelation(field, rModel)
}

func (s *builderImpl) GetFieldPlaceHolder(field model.Field) (any, *cd.Result) {
	return s.builder.GetFieldPlaceHolder(field)
}

// NewBuilder new builder
func NewBuilder(vModel model.Model, codec codec.Codec) Builder {
	return &builderImpl{
		builder: mysql.New(vModel, codec),
	}
}
