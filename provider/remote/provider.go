package remote

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
)

// Provider remote provider
type Provider struct {
	modelCache Cache
}

// New create remote provider
func New() *Provider {
	return &Provider{modelCache: NewCache()}
}

// GetObjectModel GetObjectModel
func (s *Provider) GetObjectModel(obj interface{}) (ret model.Model, err error) {
	objType := reflect.TypeOf(obj)
	if objType.Kind() == reflect.Ptr {
		objPtr, objOk := obj.(*Object)
		if !objOk {
			err = fmt.Errorf("illegal obj type")
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

		ret = objPtr.Copy()
		return
	}

	objVal, objOk := obj.(Object)
	if !objOk {
		err = fmt.Errorf("illegal obj type")
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

	ret = &objVal
	return
}

// GetValueModel GetValueModel
func (s *Provider) GetValueModel(val reflect.Value) (ret model.Model, err error) {
	objInterface := reflect.Indirect(val).Interface()
	objVal, objOK := objInterface.(ObjectValue)
	if !objOK {
		err = fmt.Errorf("illegal model value")
		return
	}

	objPtr := s.modelCache.Fetch(objVal.GetName())
	if objPtr == nil {
		err = fmt.Errorf("illegal model value, no found model, name:%s", objVal.GetName())
		return
	}

	if objPtr.GetPkgPath() != objVal.GetPkgPath() {
		err = fmt.Errorf("illegal model value, miss match pkgPath, name:%s,pkgPath:%s", objVal.GetName(), objVal.GetPkgPath())
		return
	}

	objPtr = objPtr.Copy()
	for idx := range objPtr.Items {
		item := objPtr.Items[idx]
		val, ok := objVal.Items[item.GetName()]
		if !ok {
			continue
		}

		err = item.SetValue(reflect.ValueOf(val))
		if err != nil {
			return
		}
	}
	ret = objPtr

	return
}

// GetTypeModel GetTypeModel
func (s *Provider) GetTypeModel(vType model.Type) (ret model.Model, err error) {
	objPtr := s.modelCache.Fetch(vType.GetName())
	if objPtr == nil {
		return
	}

	if objPtr.GetPkgPath() != vType.GetPkgPath() {
		err = fmt.Errorf("illegal type, pkgPath isn't match")
		return
	}

	ret = objPtr.Copy()

	return
}

// GetValueStr GetValueStr
func (s *Provider) GetValueStr(vType model.Type, vVal model.Value) (ret string, err error) {
	return getValueStr(vType, vVal, s.modelCache)
}

// GetModelDependValue GetModelDependValue
func (s *Provider) GetModelDependValue(vModel model.Model, vValue model.Value) (ret []reflect.Value, err error) {
	if vValue.IsNil() {
		return
	}

	val := reflect.Indirect(vValue.Get())
	if val.Kind() == reflect.Slice {
		for idx := 0; idx < val.Len(); idx++ {
			sliceItem := val.Index(idx)
			itemModel, itemErr := s.GetValueModel(sliceItem)
			if itemErr != nil {
				err = itemErr
				return
			}

			if itemModel.GetName() != vModel.GetName() || itemModel.GetPkgPath() != vModel.GetPkgPath() {
				err = fmt.Errorf("illegal slice model value, type:%s", val.Type().String())
				return
			}

			ret = append(ret, sliceItem)
		}
	} else if val.Kind() == reflect.Struct {
		itemModel, itemErr := s.GetValueModel(val)
		if itemErr != nil {
			err = itemErr
			return
		}

		if itemModel.GetName() != vModel.GetName() || itemModel.GetPkgPath() != vModel.GetPkgPath() {
			err = fmt.Errorf("illegal struct model value, type:%s", val.Type().String())
			return
		}

		ret = append(ret, vValue.Get())
	} else {
		err = fmt.Errorf("illegal value type, type:%s", val.Type().String())
	}

	return
}

// Reset Reset
func (s *Provider) Reset() {
	s.modelCache.Reset()
}
