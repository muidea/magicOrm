package local

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/muidea/magicOrm/model"
)

type specImpl struct {
	specVal string
}

// newSpec name[key][auto]
func newSpec(tag reflect.StructTag) (ret *specImpl, err error) {
	spec := tag.Get("orm")
	ret, err = getSpec(spec)
	return
}

func getSpec(spec string) (ret *specImpl, err error) {
	items := strings.Split(spec, "")
	if len(items) < 1 {
		err = fmt.Errorf("illegal spec value, val:%s", spec)
		return
	}

	ret = &specImpl{specVal: spec}
	return
}

// GetFieldName Name
func (s *specImpl) GetFieldName() (ret string) {
	items := strings.Split(s.specVal, " ")
	ret = items[0]

	return
}

func (s *specImpl) IsPrimaryKey() (ret bool) {
	items := strings.Split(s.specVal, " ")
	if len(items) <= 1 {
		return false
	}

	isPrimaryKey := false
	if len(items) >= 2 {
		switch items[1] {
		case "key":
			isPrimaryKey = true
		}
	}
	if len(items) >= 3 {
		switch items[2] {
		case "key":
			isPrimaryKey = true
		}
	}

	ret = isPrimaryKey
	return
}

func (s *specImpl) GetValueDeclare() model.ValueDeclare {
	if s.IsAutoIncrement() {
		return model.AutoIncrement
	}

	return model.Customer
}

// IsAutoIncrement IsAutoIncrement
func (s *specImpl) IsAutoIncrement() (ret bool) {
	items := strings.Split(s.specVal, " ")
	if len(items) <= 1 {
		return false
	}

	isAutoIncrement := false
	if len(items) >= 2 {
		switch items[1] {
		case "auto":
			isAutoIncrement = true
		}
	}
	if len(items) >= 3 {
		switch items[2] {
		case "auto":
			isAutoIncrement = true
		}
	}

	ret = isAutoIncrement
	return
}

func (s *specImpl) IsUUID() (ret bool) {
	items := strings.Split(s.specVal, " ")
	if len(items) <= 1 {
		return false
	}

	isUUID := false
	if len(items) >= 2 {
		switch items[1] {
		case "uuid":
			isUUID = true
		}
	}
	if len(items) >= 3 {
		switch items[2] {
		case "uuid":
			isUUID = true
		}
	}

	ret = isUUID
	return
}

func (s *specImpl) IsSnowFlake() (ret bool) {
	items := strings.Split(s.specVal, " ")
	if len(items) <= 1 {
		return false
	}

	isSnowFlake := false
	if len(items) >= 2 {
		switch items[1] {
		case "snowflake":
			isSnowFlake = true
		}
	}
	if len(items) >= 3 {
		switch items[2] {
		case "snowflake":
			isSnowFlake = true
		}
	}

	ret = isSnowFlake
	return
}

func (s *specImpl) copy() (ret *specImpl) {
	ret = &specImpl{specVal: s.specVal}
	return
}

func (s *specImpl) dump() (ret string) {
	return fmt.Sprintf("name=%s key=%v auto=%v, uuid=%v, snowFlake=%v",
		s.GetFieldName(), s.IsPrimaryKey(), s.IsAutoIncrement(), s.IsUUID(), s.IsSnowFlake())
}
