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
	obj := &Object{Name: s.Name, PkgPath: s.PkgPath, IsPtr: s.IsPtr, Items: []*Item{}}
	for _, val := range s.Items {
		obj.Items = append(obj.Items, &Item{Index: val.Index, Name: val.Name, Tag: val.Tag, Type: val.Type, value: val.value})
	}

	ret = obj
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
func GetObject(obj interface{}, cache Cache) (ret *Object, err error) {
	objVal := reflect.ValueOf(obj)
	if objVal.Kind() == reflect.Ptr {
		objVal = reflect.Indirect(objVal)
	}

	objType := objVal.Type()
	ret, err = Type2Object(objType, cache)
	return
}

// Type2Object Type2Object
func Type2Object(objType reflect.Type, cache Cache) (ret *Object, err error) {
	objPtr := false
	if objType.Kind() == reflect.Ptr {
		objPtr = true
		objType = objType.Elem()
	}

	if objType.Kind() != reflect.Struct {
		err = fmt.Errorf("illegal obj type, must be a struct obj, type:%s", objType.String())
		return
	}

	ret = cache.Fetch(objType.Name())
	if ret != nil {
		return
	}

	ret = &Object{}
	ret.Name = objType.Name()
	ret.PkgPath = objType.PkgPath()
	ret.IsPtr = objPtr
	ret.Items = []*Item{}

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

		itemType, itemErr := GetItemType(fType, cache)
		if itemErr != nil {
			err = itemErr
			return
		}

		fItem := &Item{Index: idx, Name: field.Name, Tag: *itemTag, Type: *itemType, value: ItemValue{}}

		ret.Items = append(ret.Items, fItem)
	}

	return
}