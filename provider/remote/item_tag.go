package remote

import (
	"fmt"
	"strings"

	"muidea.com/magicOrm/model"
)

// ItemTag ItemTag
type ItemTag struct {
	Tag string `json:"tag"`
}

// GetName Name
func (s *ItemTag) GetName() (ret string) {
	items := strings.Split(s.Tag, " ")
	ret = items[0]

	return
}

// IsPrimaryKey IsPrimaryKey
func (s *ItemTag) IsPrimaryKey() (ret bool) {
	items := strings.Split(s.Tag, " ")
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
func (s *ItemTag) IsAutoIncrement() (ret bool) {
	items := strings.Split(s.Tag, " ")
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

func (s *ItemTag) String() (ret string) {
	return fmt.Sprintf("name=%s key=%v auto=%v", s.GetName(), s.IsPrimaryKey(), s.IsAutoIncrement())
}

// Copy Copy
func (s *ItemTag) Copy() (ret model.Tag) {
	return &ItemTag{Tag: s.Tag}
}

// GetItemTag Verify Item Tag
func GetItemTag(tag string) (ret *ItemTag, err error) {
	items := strings.Split(tag, "")
	if len(items) < 1 {
		err = fmt.Errorf("illegal tag value")
		return
	}

	ret = &ItemTag{Tag: tag}

	return
}
