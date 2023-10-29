package local

import (
	"fmt"
	"reflect"
	"strings"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/util"
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
	items := strings.Split(spec, " ")
	if len(items) < 1 {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal spec value, val:%s", spec))
		return
	}

	ret = &SpecImpl{primaryKey: false, valueDeclare: model.Customer}
	ret.fieldName = items[0]
	for idx := 1; idx < len(items); idx++ {
		switch items[idx] {
		case util.Auto:
			ret.valueDeclare = model.AutoIncrement
		case util.UUID:
			ret.valueDeclare = model.UUID
		case util.SnowFlake:
			ret.valueDeclare = model.SnowFlake
		case util.DateTime:
			ret.valueDeclare = model.DateTime
		case util.Key:
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
		case "view":
			ret = append(ret, model.FullView)
		case "lite":
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
	for _, val := range s.viewDeclare {
		if val == viewSpec {
			return true
		}
	}

	return false
}

func (s *SpecImpl) copy() (ret *SpecImpl) {
	ret = &SpecImpl{
		fieldName:    s.fieldName,
		primaryKey:   s.primaryKey,
		valueDeclare: s.valueDeclare,
	}

	return
}

func (s *SpecImpl) dump() (ret string) {
	return fmt.Sprintf("name=%s key=%v value=%v", s.GetFieldName(), s.IsPrimaryKey(), s.GetValueDeclare())
}
