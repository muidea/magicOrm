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

func (s *ItemType) GetName() (ret string) {
	return
}

func (s *ItemType) GetValue() (ret int) {
	return
}

func (s *ItemType) GetPkgPath() (ret string) {
	return
}

func (s *ItemType) GetType() (ret reflect.Type) {
	return
}

func (s *ItemType) GetDepend() (ret model.FieldType) {
	return
}

func (s *ItemType) IsPtrType() (ret bool) {
	return
}

func (s *ItemType) String() (ret string) {
	return
}

func (s *ItemType) Copy() (ret model.FieldType) {
	return
}
