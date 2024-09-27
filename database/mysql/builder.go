package mysql

import (
	"github.com/muidea/magicOrm/database/common"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

// Builder Builder
type Builder struct {
	common *common.Common
}

// New create builder
func New(vModel model.Model, modelProvider provider.Provider, prefix string) *Builder {
	return &Builder{
		common: common.New(vModel, modelProvider, prefix),
	}
}
