package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database/common"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

// Builder Builder
type Builder struct {
	common.Common
}

// New create builder
func New(vModel model.Model, modelProvider provider.Provider, prefix string) *Builder {
	return &Builder{Common: common.New(vModel, modelProvider, prefix)}
}

func (s *Builder) buildModelFilter() (ret string, err *cd.Result) {
	pkField := s.GetPrimaryKeyField(nil)
	pkfVal, pkfErr := s.BuildFieldValue(pkField.GetType(), pkField.GetValue())
	if pkfErr != nil {
		err = pkfErr
		log.Errorf("buildModelFilter failed, s.EncodeValue error:%s", err.Error())
		return
	}

	pkfName := pkField.GetName()
	ret = fmt.Sprintf("`%s` = %v", pkfName, pkfVal)
	return
}
