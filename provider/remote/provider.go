package remote

import (
	"fmt"
	"reflect"

	"muidea.com/magicOrm/model"
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

		ret = objPtr
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
		err = fmt.Errorf("illegal value")
		return
	}

	objPtr := s.modelCache.Fetch(objVal.TypeName)
	if objPtr == nil {
		err = fmt.Errorf("illegal value, no found model")
		return
	}

	if objPtr.GetPkgPath() != objVal.TypeName {
		err = fmt.Errorf("illegal value, pkgPath isn't match")
		return
	}

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

	ret = objPtr
	return
}

// GetValueStr GetValueStr
func (s *Provider) GetValueStr(vType model.Type, vVal model.Value) (ret string, err error) {
	return
}

// GetModelDependValue GetModelDependValue
func (s *Provider) GetModelDependValue(vModel model.Model, vVal model.Value) (ret []reflect.Value, err error) {
	return
}

// Reset Reset
func (s *Provider) Reset() {
	s.modelCache.Reset()
}
