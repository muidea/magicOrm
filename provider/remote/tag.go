package remote

import (
	"fmt"
	"strings"
)

// TagImpl TagImpl
type TagImpl struct {
	Value string `json:"value"`
}

// GetName Name
func (s *TagImpl) GetName() (ret string) {
	items := strings.Split(s.Value, " ")
	ret = items[0]

	return
}

// IsPrimaryKey IsPrimaryKey
func (s *TagImpl) IsPrimaryKey() (ret bool) {
	items := strings.Split(s.Value, " ")
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
func (s *TagImpl) IsAutoIncrement() (ret bool) {
	items := strings.Split(s.Value, " ")
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

func (s *TagImpl) String() (ret string) {
	return fmt.Sprintf("name=%s key=%v auto=%v", s.GetName(), s.IsPrimaryKey(), s.IsAutoIncrement())
}

// Copy Copy
func (s *TagImpl) Copy() (ret *TagImpl) {
	return &TagImpl{Value: s.Value}
}

// GetTag get Item Value
func GetTag(tag string) (ret *TagImpl, err error) {
	items := strings.Split(tag, "")
	if len(items) < 1 {
		err = fmt.Errorf("illegal tag value")
		return
	}

	ret = &TagImpl{Value: tag}

	return
}
