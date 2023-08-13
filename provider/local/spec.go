package local

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/util"
)

type specImpl struct {
	fieldName    string
	primaryKey   bool
	valueDeclare model.ValueDeclare
}

var emptySpec = specImpl{primaryKey: false, valueDeclare: model.Customer}

// newSpec name[key][auto]
func newSpec(tag reflect.StructTag) (ret *specImpl, err error) {
	spec := tag.Get("orm")
	ret, err = getSpec(spec)
	return
}

func getSpec(spec string) (ret *specImpl, err error) {
	items := strings.Split(spec, " ")
	if len(items) < 1 {
		err = fmt.Errorf("illegal spec value, val:%s", spec)
		return
	}

	ret = &specImpl{primaryKey: false, valueDeclare: model.Customer}
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
func (s *specImpl) GetFieldName() string {
	return s.fieldName
}

func (s *specImpl) IsPrimaryKey() bool {
	return s.primaryKey
}

func (s *specImpl) GetValueDeclare() model.ValueDeclare {
	return s.valueDeclare
}

func (s *specImpl) copy() (ret *specImpl) {
	ret = &specImpl{
		fieldName:    s.fieldName,
		primaryKey:   s.primaryKey,
		valueDeclare: s.valueDeclare,
	}

	return
}

func (s *specImpl) dump() (ret string) {
	return fmt.Sprintf("name=%s key=%v value=%v", s.GetFieldName(), s.IsPrimaryKey(), s.GetValueDeclare())
}
