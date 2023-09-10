package local

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/util"
)

type SpecImpl struct {
	fieldName    string
	primaryKey   bool
	valueDeclare model.ValueDeclare
}

var emptySpec = SpecImpl{primaryKey: false, valueDeclare: model.Customer}

// NewSpec name[key][auto]
func NewSpec(tag reflect.StructTag) (ret *SpecImpl, err error) {
	spec := tag.Get("orm")
	ret, err = getSpec(spec)
	return
}

func getSpec(spec string) (ret *SpecImpl, err error) {
	items := strings.Split(spec, " ")
	if len(items) < 1 {
		err = fmt.Errorf("illegal spec value, val:%s", spec)
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
