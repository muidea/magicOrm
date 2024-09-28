package mysql

import (
	"github.com/muidea/magicOrm/database/context"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

// Builder Builder
type Builder struct {
	common *context.Context
}

// New create builder
func New(vModel model.Model, modelProvider provider.Provider, prefix string) *Builder {
	return &Builder{
		common: context.New(vModel, modelProvider, prefix),
	}
}
