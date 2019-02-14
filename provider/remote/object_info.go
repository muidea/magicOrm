package remote

import (
	"fmt"
	"reflect"

	"muidea.com/magicOrm/model"
)

// Object Object
type Object struct {
	Name    string  `json:"name"`
	PkgPath string  `json:"pkgPath"`
	IsPtr   bool    `json:"isPtr"`
	Items   []*Item `json:"items"`
}

// GetName GetName
func (s *Object) GetName() (ret string) {
	ret = s.Name
	return
}

// GetPkgPath GetPkgPath
func (s *Object) GetPkgPath() (ret string) {
	ret = s.PkgPath
	return
}

// GetFields GetFields
func (s *Object) GetFields() (ret model.Fields) {
	for _, val := range s.Items {
		ret = append(ret, val)
	}

	return
}

// SetFieldValue SetFieldValue
func (s *Object) SetFieldValue(idx int, val reflect.Value) (err error) {
	for _, item := range s.Items {
		if item.Index == idx {
			err = item.SetValue(val)
			return
		}
	}

	return
}

// UpdateFieldValue UpdateFieldValue
func (s *Object) UpdateFieldValue(name string, val reflect.Value) (err error) {
	for _, item := range s.Items {
		if item.Name == name {
			err = item.SetValue(val)
			return
		}
	}

	return
}

// GetPrimaryField GetPrimaryField
func (s *Object) GetPrimaryField() (ret model.Field) {
	for _, v := range s.Items {
		if v.IsPrimary() {
			ret = v
			return
		}
	}

	return
}

// GetDependField GetDependField
func (s *Object) GetDependField() (ret []model.Field) {
	for _, v := range s.Items {
		if v.Type.GetDepend() != nil {
			ret = append(ret, v)
		}
	}

	return
}

// Copy Copy
func (s *Object) Copy() (ret model.Model) {
	info := &Object{Name: s.Name, PkgPath: s.PkgPath, IsPtr: s.IsPtr, Items: []*Item{}}
	for _, val := range s.Items {
		info.Items = append(info.Items, &Item{Index: val.Index, Name: val.Name, Tag: val.Tag, Type: val.Type, value: val.value})
	}

	ret = info
	return
}

// Interface Interface
func (s *Object) Interface() (ret reflect.Value) {
	return
}

// Dump Dump
func (s *Object) Dump() {

}

// GetObject GetObject
func GetObject(obj interface{}) (info *Object, err error) {
	objVal := reflect.ValueOf(obj)
	if objVal.Kind() == reflect.Ptr {
		objVal = reflect.Indirect(objVal)
	}

	objType := objVal.Type()
	info, err = Type2Info(objType)
	return
}

// Type2Info Type2Info
func Type2Info(objType reflect.Type) (info *Object, err error) {
	objPtr := false
	if objType.Kind() == reflect.Ptr {
		objPtr = true
		objType = objType.Elem()
	}

	if objType.Kind() != reflect.Struct {
		err = fmt.Errorf("illegal obj type, must be a struct obj, type:%s", objType.String())
		return
	}

	info = &Object{}
	info.Name = objType.Name()
	info.PkgPath = objType.PkgPath()
	info.IsPtr = objPtr
	info.Items = []*Item{}

	fieldNum := objType.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		field := objType.Field(idx)
		fType := field.Type
		fTag := field.Tag.Get("orm")

		itemTag, itemErr := GetItemTag(fTag)
		if itemErr != nil {
			err = itemErr
			return
		}

		itemType, itemErr := GetItemType(fType)
		if itemErr != nil {
			err = itemErr
			return
		}

		fItem := &Item{Index: idx, Name: field.Name, Tag: *itemTag, Type: *itemType, value: ItemValue{}}

		info.Items = append(info.Items, fItem)
	}

	return
}
