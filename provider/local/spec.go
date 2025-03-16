package local

import (
	"reflect"
	"strings"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/utils"
)

const (
	ormTag  = "orm"
	viewTag = "view"
)

type SpecImpl struct {
	fieldName    string
	primaryKey   bool
	valueDeclare model.ValueDeclare
	viewDeclare  []model.ViewDeclare
}

var emptySpec = SpecImpl{primaryKey: false, valueDeclare: model.Customer}

// NewSpec name[key][auto]
func NewSpec(tag reflect.StructTag) (ret *SpecImpl, err *cd.Result) {
	ormSpec := tag.Get(ormTag)
	ret, err = getOrmSpec(ormSpec)
	if err != nil {
		return
	}
	viewSpec := tag.Get(viewTag)
	ret.viewDeclare = getViewItems(viewSpec)
	return
}

func getOrmSpec(spec string) (ret *SpecImpl, err *cd.Result) {
	items := strings.Split(strings.TrimSpace(spec), " ")
	ret = &SpecImpl{primaryKey: false, valueDeclare: model.Customer}
	ret.fieldName = items[0]
	for idx := 1; idx < len(items); idx++ {
		switch items[idx] {
		case utils.Auto:
			ret.valueDeclare = model.AutoIncrement
		case utils.UUID:
			ret.valueDeclare = model.UUID
		case utils.SnowFlake:
			ret.valueDeclare = model.SnowFlake
		case utils.DateTime:
			ret.valueDeclare = model.DateTime
		case utils.Key:
			ret.primaryKey = true
		}
	}

	return
}

func getViewItems(spec string) (ret []model.ViewDeclare) {
	ret = []model.ViewDeclare{}
	items := strings.Split(spec, ",")
	for _, sv := range items {
		switch sv {
		case model.DetailView:
			ret = append(ret, model.DetailView)
		case model.LiteView:
			ret = append(ret, model.LiteView)
		}
	}
	return
}

// GetFieldName Name
func (s *SpecImpl) GetFieldName() string {
	return s.fieldName
}

func (s *SpecImpl) IsPrimaryKey() bool {
	return s.primaryKey
}

func (s *SpecImpl) GetValueDeclare() model.ValueDeclare {
	return s.valueDeclare
}

func (s *SpecImpl) EnableView(viewSpec model.ViewDeclare) bool {
	if viewSpec == model.MetaView {
		return true
	}

	for _, val := range s.viewDeclare {
		if val == viewSpec {
			return true
		}
	}

	return false
}

func (s *SpecImpl) GetDefaultValue() any {
	return nil
}
