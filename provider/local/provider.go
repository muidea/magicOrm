package local

import (
	"fmt"
	"reflect"

	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/util"
)

// Provider local provider
type Provider struct {
	modelCache Cache
}

// New create local provider
func New() *Provider {
	return &Provider{modelCache: NewCache()}
}

// GetObjectModel GetObjectModel
func (s *Provider) GetObjectModel(objPtr interface{}) (ret model.Model, err error) {
	return getObjectModel(objPtr, s.modelCache)
}

// GetTypeModel GetTypeModel
func (s *Provider) GetTypeModel(modelType reflect.Type) (ret model.Model, err error) {
	typeImpl, typeErr := newType(modelType)
	if typeErr != nil {
		err = typeErr
		return
	}

	if util.IsBasicType(typeImpl.GetValue()) {
		return
	}

	if util.IsSliceType(typeImpl.GetValue()) {
		rawType := typeImpl.GetType().Elem()
		typeImpl, typeErr = newType(rawType)
		if typeErr != nil {
			err = typeErr
			return
		}

		if util.IsBasicType(typeImpl.GetValue()) {
			return
		}

		return getTypeModel(rawType, s.modelCache)
	}

	return getTypeModel(modelType, s.modelCache)
}

// GetValueModel GetValueModel
func (s *Provider) GetValueModel(modelVal reflect.Value) (ret model.Model, err error) {
	return getValueModel(modelVal, s.modelCache)
}

// GetValueStr GetValueStr
func (s *Provider) GetValueStr(vType model.Type, vVal model.Value) (ret string, err error) {
	return getValueStr(vType, vVal, s.modelCache)
}

// GetSliceModelValueStr GetSliceModelValueStr
func (s *Provider) GetSliceModelValueStr(vType model.Model, vVal model.Value) (ret []string, err error) {
	val := reflect.Indirect(vVal.Get())
	if val.Kind() != reflect.Slice {
		err = fmt.Errorf("illegal slice model value, type:%s", val.Type().String())
		return
	}

	for idx := 0; idx < val.Len(); idx++ {
		item := reflect.Indirect(val.Index(idx))
		itemType := item.Type()
		if itemType.Name() != vType.GetName() || itemType.PkgPath() != vType.GetPkgPath() {
			err = fmt.Errorf("illegal slice model value, type:%s", val.Type().String())
			return
		}

		val, valErr := getStructValueStr(item, s.modelCache)
		if valErr != nil {
			err = valErr
			return
		}

		ret = append(ret, val)
	}

	return
}

// Reset Reset
func (s *Provider) Reset() {
	s.modelCache.Reset()
}
