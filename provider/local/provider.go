package local

import (
	"fmt"
	"log"
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
	modelVal := reflect.ValueOf(objPtr)

	modelImpl, modelErr := getValueModel(modelVal, s.modelCache)
	if modelErr != nil {
		err = modelErr
		log.Printf("getValueModel failed, err:%s", err.Error())
		return
	}

	ret = modelImpl
	return
}

// GetTypeModel GetTypeModel
func (s *Provider) GetTypeModel(vType model.Type) (ret model.Model, err error) {
	if util.IsBasicType(vType.GetValue()) {
		return
	}

	if util.IsSliceType(vType.GetValue()) {
		rawType := vType.Elem()
		if util.IsBasicType(rawType.GetValue()) {
			return
		}

		modelImpl, modelErr := getTypeModel(rawType, s.modelCache)
		if modelErr != nil {
			err = modelErr
			return
		}

		ret = modelImpl
		return
	}

	modelImpl, modelErr := getTypeModel(vType, s.modelCache)
	if modelErr != nil {
		err = modelErr
		return
	}

	ret = modelImpl
	return
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
	if vValue.IsNil() {
		return
	}

	val := reflect.Indirect(vValue.Get())
	if val.Kind() == reflect.Slice {
		for idx := 0; idx < val.Len(); idx++ {
			sliceItem := val.Index(idx)
			rawType, rawErr := newType(sliceItem.Type())
			if rawErr != nil {
				err = rawErr
				return
			}

			if rawType.GetName() != vModel.GetName() || rawType.GetPkgPath() != vModel.GetPkgPath() {
				err = fmt.Errorf("illegal slice model value, type:%s", val.Type().String())
				return
			}

			ret = append(ret, sliceItem)
		}
	} else if val.Kind() == reflect.Struct {
		rawType, rawErr := newType(val.Type())
		if rawErr != nil {
			err = rawErr
			return
		}

		if rawType.GetName() != vModel.GetName() || rawType.GetPkgPath() != vModel.GetPkgPath() {
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
