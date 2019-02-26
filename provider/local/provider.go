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

		return getTypeModel(typeImpl.GetType(), s.modelCache)
	}

	return getTypeModel(typeImpl.GetType(), s.modelCache)
}

// GetValueModel GetValueModel
func (s *Provider) GetValueModel(modelVal reflect.Value) (ret model.Model, err error) {
	return getValueModel(modelVal, s.modelCache)
}

// GetValueStr GetValueStr
func (s *Provider) GetValueStr(vType model.Type, vValue model.Value) (ret string, err error) {
	return getValueStr(vType, vValue, s.modelCache)
}

// GetModelDependValue GetModelDependValue
func (s *Provider) GetModelDependValue(vModel model.Model, vValue model.Value) (ret []reflect.Value, err error) {
	val := reflect.Indirect(vValue.Get())
	if val.Kind() == reflect.Slice {
		for idx := 0; idx < val.Len(); idx++ {
			item := reflect.Indirect(val.Index(idx))
			itemType := item.Type()
			if itemType.Name() != vModel.GetName() || itemType.PkgPath() != vModel.GetPkgPath() {
				err = fmt.Errorf("illegal slice model value, type:%s", val.Type().String())
				return
			}

			ret = append(ret, item)
		}
	} else if val.Kind() == reflect.Struct {
		valType := val.Type()
		if valType.Name() != vModel.GetName() || valType.PkgPath() != vModel.GetPkgPath() {
			err = fmt.Errorf("illegal struct model value, type:%s", val.Type().String())
			return
		}

		ret = append(ret, val)
	} else {
		err = fmt.Errorf("illegal value type, type:%s", val.Type().String())
	}

	return
}

// Reset Reset
func (s *Provider) Reset() {
	s.modelCache.Reset()
}
