package remote

import (
	"reflect"

	"muidea.com/magicOrm/model"
)

type ItemType struct {
	TypeName  string `json:"name"`
	TypeValue int    `json:"value"`
	IsPtr     bool   `json:"isPtr"`
	PkgPath   string `json:"pkgPath"`
	Depend    *Info  `json:"depend"`
}

func (s *ItemType) Name() (ret string) {
	return
}

func (s *ItemType) Value() (ret int) {
	return
}

func (s *ItemType) IsPtr() (ret bool) {
	return
}

func (s *ItemType) PkgPath() (ret string) {
	return
}

func (s *ItemType) String() (ret string) {
	return
}

func (s *ItemType) Type() (ret reflect.Type) {
	return
}

func (s *ItemType) Depend() (ret model.FieldType) {
	return
}

func (s *ItemType) Copy() (ret model.FieldType) {
	return
}
