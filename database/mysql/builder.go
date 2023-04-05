package mysql

import (
	"fmt"
	"github.com/muidea/magicOrm/database/common"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

// Builder Builder
type Builder struct {
	common.Common
}

// New create builder
func New(vModel model.Model, modelProvider provider.Provider) *Builder {
	return &Builder{Common: common.New(vModel, modelProvider)}
}

func (s *Builder) buildModelFilter() (ret string, err error) {
	pkField := s.GetPrimaryKeyField()
	pkfVal, pkfErr := s.EncodeValue(pkField.GetValue(), pkField.GetType())
	if pkfErr != nil {
		err = pkfErr
		return
	}

	pkfTag := pkField.GetTag().GetName()
	ret = fmt.Sprintf("`%s`=%v", pkfTag, pkfVal)
	return
}
