package remote

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/util"
)

// Provider remote provider
type Provider struct {
	modelCache Cache
}

// New create remote provider
func New() *Provider {
	return &Provider{modelCache: NewCache()}
}

// RegisterObjectModel RegisterObjectModel
func (s *Provider) RegisterObjectModel(obj interface{}) (err error) {
	objType := reflect.TypeOf(obj)
	if objType.Kind() == reflect.Ptr {
		objPtr, objOk := obj.(*Object)
		if !objOk {
			err = fmt.Errorf("illegal obj, isn't Object ptr")
			return
		}

		preObj := s.modelCache.Fetch(objPtr.GetName())
		if preObj != nil {
			if objPtr.GetPkgPath() != preObj.GetPkgPath() {
				err = fmt.Errorf("illegal object, pkgPath isn't match")
				return
			}
		} else {
			s.modelCache.Put(objPtr.GetName(), objPtr)
		}

		return
	}

	objVal, objOk := obj.(Object)
	if !objOk {
		err = fmt.Errorf("illegal obj, isn't Object")
		return
	}

	preObj := s.modelCache.Fetch(objVal.GetName())
	if preObj != nil {
		if objVal.GetPkgPath() != preObj.GetPkgPath() {
			err = fmt.Errorf("illegal object, pkgPath isn't match")
			return
		}
	} else {
		s.modelCache.Put(objVal.GetName(), &objVal)
	}

	return
}

// UnregisterModel register model
func (s *Provider) UnregisterModel(obj interface{}) {
	objType := reflect.TypeOf(obj)
	if objType.Kind() == reflect.Ptr {
		objPtr, objOk := obj.(*Object)
		if !objOk {
			return
		}

		preObj := s.modelCache.Fetch(objPtr.GetName())
		if preObj != nil {
			if objPtr.GetPkgPath() != preObj.GetPkgPath() {
				return
			}

			s.modelCache.Remove(preObj.GetName())
		}

		return
	}

	objVal, objOk := obj.(Object)
	if !objOk {
		return
	}

	preObj := s.modelCache.Fetch(objVal.GetName())
	if preObj != nil {
		if objVal.GetPkgPath() != preObj.GetPkgPath() {
			return
		}

		s.modelCache.Remove(preObj.GetName())
	}

	return
}

// GetObjectModel GetObjectModel
func (s *Provider) GetObjectModel(obj interface{}) (ret model.Model, err error) {
	objType := reflect.TypeOf(obj)
	if objType.Kind() == reflect.Ptr {
		objPtr, objOk := obj.(*Object)
		if !objOk {
			err = fmt.Errorf("illegal obj type, obj type:%s", objType.String())
			return
		}

		preObj := s.modelCache.Fetch(objPtr.GetName())
		if preObj != nil {
			if objPtr.GetPkgPath() != preObj.GetPkgPath() {
				err = fmt.Errorf("illegal object, pkgPath isn't match")
				return
			}
		} else {
			err = fmt.Errorf("can't find obj model, objName:%s, objPkgPath:%s", objPtr.GetName(), objPtr.GetPkgPath())
			return
		}

		objModel := preObj.Copy()
		objModel.IsPtr = true
		ret = objModel

		return
	}

	objVal, objOk := obj.(Object)
	if !objOk {
		err = fmt.Errorf("illegal obj type, obj type:%s", objType.String())
		return
	}

	preObj := s.modelCache.Fetch(objVal.GetName())
	if preObj != nil {
		if objVal.GetPkgPath() != preObj.GetPkgPath() {
			err = fmt.Errorf("illegal object, pkgPath isn't match")
			return
		}
	} else {
		err = fmt.Errorf("can't find obj model, objName:%s, objPkgPath:%s", objVal.GetName(), objVal.GetPkgPath())
		return
	}

	objModel := preObj.Copy()
	objModel.IsPtr = false
	ret = objModel

	return
}

// GetValueModel GetValueModel
func (s *Provider) GetValueModel(val reflect.Value) (ret model.Model, err error) {
	objImpl, objErr := getValueModel(val, s.modelCache)
	if objErr != nil {
		err = objErr
		return
	}

	ret = objImpl
	return
}

// GetSliceValueModel GetSliceValueModel
func (s *Provider) GetSliceValueModel(modelVal reflect.Value) (ret model.Model, err error) {
	return
}

// GetTypeModel GetTypeModel
func (s *Provider) GetTypeModel(vType model.Type) (ret model.Model, err error) {
	depend := vType.Depend()
	if depend == nil || util.IsBasicType(depend.GetValue()) {
		return
	}
	vType = depend

	typeImpl, typeErr := getTypeMode(vType, s.modelCache)
	if typeErr != nil {
		err = typeErr
		return
	}

	if typeImpl != nil {
		ret = typeImpl
	}

	return
}

// GetValueStr GetValueStr
func (s *Provider) GetValueStr(vType model.Type, vVal model.Value) (ret string, err error) {
	ret, err = getValueStr(vType, vVal, s.modelCache)
	return
}

// GetModelDependValue GetModelDependValue
func (s *Provider) GetModelDependValue(vModel model.Model, vValue model.Value) (ret []reflect.Value, err error) {
	if vValue.IsNil() {
		return
	}

	val := vValue.Get()
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	val = reflect.Indirect(val)
	if val.Kind() == reflect.Slice {
		for idx := 0; idx < val.Len(); idx++ {
			sliceVal := val.Index(idx)
			itemModel, itemErr := getValueModel(sliceVal, s.modelCache)
			if itemErr != nil {
				err = itemErr
				return
			}

			if itemModel.GetName() != vModel.GetName() || itemModel.GetPkgPath() != vModel.GetPkgPath() {
				err = fmt.Errorf("illegal slice model value, item type name:%s, expect type:%s", itemModel.GetName(), vModel.GetName())
				return
			}

			ret = append(ret, sliceVal)
		}

		return
	}

	itemModel, itemErr := getValueModel(val, s.modelCache)
	if itemErr != nil {
		err = itemErr
		return
	}

	if itemModel.GetName() != vModel.GetName() || itemModel.GetPkgPath() != vModel.GetPkgPath() {
		err = fmt.Errorf("illegal struct value, item type name:%s, expect type:%s", itemModel.GetName(), vModel.GetName())
		return
	}

	ret = append(ret, val)

	return
}

// Reset Reset
func (s *Provider) Reset() {
	s.modelCache.Reset()
}

var _referenceVal ObjectValue
var _referenceType = reflect.TypeOf(_referenceVal)

func getValueModel(val reflect.Value, cache Cache) (ret *Object, err error) {
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}

	val = reflect.Indirect(val)
	if val.Type().String() != _referenceType.String() {
		err = fmt.Errorf("illegal model value, value type:%s", val.Type().String())
		return
	}

	nameVal := val.FieldByName("TypeName")
	pkgVal := val.FieldByName("PkgPath")
	isPtr := val.FieldByName("IsPtrFlag")
	itemsVal := val.FieldByName("Items")

	objPtr := cache.Fetch(nameVal.String())
	if objPtr == nil {
		err = fmt.Errorf("illegal model value, no found model, name:%s", nameVal.String())
		return
	}

	if objPtr.GetPkgPath() != pkgVal.String() {
		err = fmt.Errorf("illegal model value, miss match pkgPath, name:%s,pkgPath:%s", nameVal.String(), pkgVal.String())
		return
	}

	if itemsVal.Len() > 0 {
		offset := 0
		objPtr = objPtr.Copy()
		for idx := range objPtr.Items {
			item := objPtr.Items[idx]
			itemVal := itemsVal.Index(offset)
			itemName := itemVal.FieldByName("Name").String()
			if item.GetName() != itemName {
				continue
			}

			offset++
			itemValue := itemVal.FieldByName("Value")
			if !util.IsNil(itemValue) {
				err = item.SetValue(itemValue)
				if err != nil {
					return
				}
			}
		}
	}

	ret = objPtr
	ret.IsPtr = isPtr.Bool()

	return
}

func getTypeMode(vType model.Type, cache Cache) (ret *Object, err error) {
	isPtr := vType.IsPtrType()

	objPtr := cache.Fetch(vType.GetName())
	if objPtr == nil {
		return
	}

	if objPtr.GetPkgPath() != vType.GetPkgPath() {
		err = fmt.Errorf("illegal type, pkgPath isn't match, type name:%s, pkgPath:%s", vType.GetName(), vType.GetPkgPath())
		return
	}

	ret = objPtr.Copy()
	ret.IsPtr = isPtr

	return
}

// getValueStr get value str
func getValueStr(vType model.Type, vVal model.Value, cache Cache) (ret string, err error) {
	if vVal.IsNil() {
		err = fmt.Errorf("invalid value")
		return
	}

	switch vType.GetValue() {
	case util.TypeBooleanField:
		ret, err = helper.EncodeBoolValue(vVal.Get())
	case util.TypeBitField, util.TypeSmallIntegerField, util.TypeInteger32Field, util.TypeIntegerField, util.TypeBigIntegerField:
		ret, err = helper.EncodeIntValue(vVal.Get())
	case util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveInteger32Field, util.TypePositiveIntegerField, util.TypePositiveBigIntegerField:
		ret, err = helper.EncodeUintValue(vVal.Get())
	case util.TypeFloatField, util.TypeDoubleField:
		ret, err = helper.EncodeFloatValue(vVal.Get())
	case util.TypeStringField:
		strRet, strErr := helper.EncodeStringValue(vVal.Get())
		if strErr != nil {
			err = strErr
			return
		}
		ret = fmt.Sprintf("'%s'", strRet)
	case util.TypeSliceField:
		strRet, strErr := helper.EncodeSliceValue(vVal.Get())
		if strErr != nil {
			err = strErr
			return
		}
		ret = fmt.Sprintf("'%s'", strRet)
	case util.TypeDateTimeField:
		strRet, strErr := helper.EncodeStringValue(vVal.Get())
		if strErr != nil {
			err = strErr
			return
		}
		ret = fmt.Sprintf("'%s'", strRet)
	case util.TypeStructField:
		valModel, valErr := getValueModel(vVal.Get(), cache)
		if valErr != nil {
			err = valErr
			return
		}

		pkField := valModel.GetPrimaryField()
		if pkField == nil {
			err = fmt.Errorf("illegal model struct")
			return
		}
		ret, err = getValueStr(pkField.GetType(), pkField.GetValue(), cache)
	default:
		err = fmt.Errorf("illegal value type, type:%v", vType.GetValue())
	}

	return
}
