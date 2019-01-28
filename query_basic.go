package orm

import (
	"muidea.com/magicOrm/model"
)

type basicValue struct {
	value model.FieldValue
}

func (s *basicValue) String() (ret string, err error) {
	ret, err = s.value.ValueStr()
	return
}
