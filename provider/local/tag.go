package local

import (
	"fmt"
	"strings"
)

type tagImpl struct {
	tagVal string
}

// newTag name[key][auto]
func newTag(val string) (ret *tagImpl, err error) {
	items := strings.Split(val, " ")
	if len(items) < 1 {
		err = fmt.Errorf("illegal tagVal value, value:%s", val)
		return
	}

	ret = &tagImpl{tagVal: val}
	return
}

// GetName Name
func (s *tagImpl) GetName() (ret string) {
	items := strings.Split(s.tagVal, " ")
	ret = items[0]

	return
}

func (s *tagImpl) IsPrimaryKey() (ret bool) {
	items := strings.Split(s.tagVal, " ")
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

// IsAutoIncrement IsAutoIncrement
func (s *tagImpl) IsAutoIncrement() (ret bool) {
	items := strings.Split(s.tagVal, " ")
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

func (s *tagImpl) copy() (ret *tagImpl) {
	ret = &tagImpl{tagVal: s.tagVal}
	return
}

func (s *tagImpl) dump() (ret string) {
	return fmt.Sprintf("name=%s key=%v auto=%v", s.GetName(), s.IsPrimaryKey(), s.IsAutoIncrement())
}
