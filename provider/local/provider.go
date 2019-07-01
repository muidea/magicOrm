package local

import (
	"fmt"
	"log"
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

// Provider local provider
type Provider struct {
	owner      string
	modelCache Cache
}

// New create local provider
func New(owner string) *Provider {
	return &Provider{owner: owner, modelCache: NewCache()}
}

// RegisterModel RegisterObjectModel
func (s *Provider) RegisterModel(entity interface{}) (err error) {
	entityType := reflect.TypeOf(entity)
	typeImpl, typeErr := newType(entityType)
	if typeErr != nil {
		err = typeErr
		return
	}

	err = registerModel(typeImpl.GetType(), s.modelCache)
	return
}

// UnregisterModel register model
func (s *Provider) UnregisterModel(entity interface{}) {
	entityType := reflect.TypeOf(entity)
	typeImpl, typeErr := newType(entityType)
	if typeErr == nil {
		cur := s.modelCache.Fetch(typeImpl.GetName())
		if cur != nil && cur.GetPkgPath() == typeImpl.GetPkgPath() {
			s.modelCache.Remove(typeImpl.GetName())
		}
	}

	return
}

// GetEntityModel GetEntityModel
func (s *Provider) GetEntityModel(objPtr interface{}) (ret model.Model, err error) {
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
	depend := vType.Depend()
	if depend == nil {
		return
	}
	if util.IsBasicType(depend.GetValue()) {
		return
	}
	vType = depend

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

// GetSliceValueModel GetSliceValueModel
func (s *Provider) GetSliceValueModel(sliceVal reflect.Value) (retModel model.Model, retVal reflect.Value, retErr error) {
	return getSliceValueModel(sliceVal, s.modelCache)
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

// Owner owner
func (s *Provider) Owner() string {
	return s.owner
}

// Reset Reset
func (s *Provider) Reset() {
	s.modelCache.Reset()
}
