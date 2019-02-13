package remote

import (
	"fmt"
	"reflect"

	"muidea.com/magicOrm/model"
)

// Info Info
type Info struct {
	Name    string  `json:"name"`
	PkgPath string  `json:"pkgPath"`
	IsPtr   bool    `json:"isPtr"`
	Items   []*Item `json:"items"`
}

// GetName GetName
func (s *Info) GetName() (ret string) {
	ret = s.Name
	return
}

// GetPkgPath GetPkgPath
func (s *Info) GetPkgPath() (ret string) {
	ret = s.PkgPath
	return
}

// GetFields GetFields
func (s *Info) GetFields() (ret model.Fields) {
	for _, val := range s.Items {
		ret = append(ret, val)
	}

	return
}

// SetFieldValue SetFieldValue
func (s *Info) SetFieldValue(idx int, val reflect.Value) (err error) {
	for _, item := range s.Items {
		if item.Index == idx {
			err = item.SetValue(val)
			return
		}
	}

	return
}

// UpdateFieldValue UpdateFieldValue
func (s *Info) UpdateFieldValue(name string, val reflect.Value) (err error) {
	for _, item := range s.Items {
		if item.Name == name {
			err = item.SetValue(val)
			return
		}
	}

	return
}

// GetPrimaryField GetPrimaryField
func (s *Info) GetPrimaryField() (ret model.Field) {
	for _, v := range s.Items {
		if v.IsPrimary() {
			ret = v
			return
		}
	}

	return
}

// GetDependField GetDependField
func (s *Info) GetDependField() (ret []model.Field) {
	for _, v := range s.Items {
		if v.Type.GetDepend() != nil {
			ret = append(ret, v)
		}
	}

	return
}

// Copy Copy
func (s *Info) Copy() (ret model.Model) {
	info := &Info{Name: s.Name, PkgPath: s.PkgPath, IsPtr: s.IsPtr, Items: []*Item{}}
	for _, val := range s.Items {
		info.Items = append(info.Items, &Item{Index: val.Index, Name: val.Name, Tag: val.Tag, Type: val.Type, value: val.value})
	}

	ret = info
	return
}

// Interface Interface
func (s *Info) Interface() (ret reflect.Value) {
	return
}

// Dump Dump
func (s *Info) Dump() {

}

// GetInfo GetInfo
func GetInfo(obj interface{}) (info *Info, err error) {
	objVal := reflect.ValueOf(obj)
	if objVal.Kind() == reflect.Ptr {
		objVal = reflect.Indirect(objVal)
	}

	objType := objVal.Type()
	info, err = Type2Info(objType)
	return
}

// Type2Info Type2Info
func Type2Info(objType reflect.Type) (info *Info, err error) {
	objPtr := false
	if objType.Kind() == reflect.Ptr {
		objPtr = true
		objType = objType.Elem()
	}

	if objType.Kind() != reflect.Struct {
		err = fmt.Errorf("illegal obj type, must be a struct obj, type:%s", objType.String())
		return
	}

	info = &Info{}
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

// AssignValue Assign Value
func (s *Info) AssignValue(val *Value) (err error) {
	if s.Name != val.TypeName || s.PkgPath != val.PkgPath {
		err = fmt.Errorf("illegal info value")
		return
	}

	for _, v := range s.Items {
		itemVal, ok := val.Items[v.Name]
		if !ok {
			continue
		}
		if itemVal == nil {
			continue
		}

		valErr := v.SetVal(itemVal)
		if valErr != nil {
			err = valErr
			return
		}
	}

	return
}

// GetPrimaryItem GetPrimaryItem
func (s *Info) GetPrimaryItem() (ret *Item, err error) {
	for _, v := range s.Items {
		if v.IsPrimary() {
			ret = v
			return
		}
	}

	err = fmt.Errorf("no defined primary item")
	return
}
