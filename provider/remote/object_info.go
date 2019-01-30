package object

import (
	"fmt"
	"reflect"

	"muidea.com/magicOrm/util"
)

// Info Info
type Info struct {
	Name    string  `json:"name"`
	PkgPath string  `json:"pkgPath"`
	IsPtr   bool    `json:"isPtr"`
	Items   []*Item `json:"items"`
}

// GetInfo GetInfo
func GetInfo(obj interface{}) (info Info, err error) {
	objVal := reflect.ValueOf(obj)
	if objVal.Kind() == reflect.Ptr {
		objVal = reflect.Indirect(objVal)
	}

	objType := objVal.Type()
	info, err = Type2Info(objType)
	return
}

// Type2Info Type2Info
func Type2Info(objType reflect.Type) (info Info, err error) {
	objPtr := false
	if objType.Kind() == reflect.Ptr {
		objPtr = true
		objType = objType.Elem()
	}

	if objType.Kind() != reflect.Struct {
		err = fmt.Errorf("illegal obj type, must be a struct obj, type:%s", objType.String())
		return
	}

	info.Name = objType.Name()
	info.PkgPath = objType.PkgPath()
	info.IsPtr = objPtr
	info.Items = []*Item{}

	fieldNum := objType.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		sf := objType.Field(idx)
		fType := sf.Type

		fPtr := false
		if fType.Kind() == reflect.Ptr {
			fType = fType.Elem()
			fPtr = true
		}

		ftVal, ftErr := util.GetTypeValueEnum(fType)
		if ftErr != nil {
			err = ftErr
			return
		}

		fItem := &Item{Name: sf.Name, Tag: sf.Tag.Get("orm"), Type: ftVal, IsPtr: fPtr}
		if util.IsStructType(ftVal) {
			modelInfo, structErr := Type2Info(fType)
			if structErr != nil {
				err = structErr
				return
			}

			fItem.DependInfo = &modelInfo
		}

		if util.IsSliceType(ftVal) {
			slicePtr := false
			fType = fType.Elem()
			if fType.Kind() == reflect.Ptr {
				fType = fType.Elem()
				slicePtr = true
			}
			ftVal, ftErr = util.GetTypeValueEnum(fType)
			if ftErr != nil {
				err = ftErr
				return
			}
			if util.IsSliceType(ftVal) {
				err = fmt.Errorf("illegal slice type, type:%s", fType.String())
				return
			}

			if util.IsStructType(ftVal) {
				sliceItem, sliceErr := Type2Info(fType)
				if sliceErr != nil {
					err = sliceErr
					return
				}
				sliceItem.IsPtr = slicePtr
				fItem.DependInfo = &sliceItem
			}
		}

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
