package local

import (
	"reflect"

	"muidea.com/magicOrm/model"
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

// Reset Reset
func (s *Provider) Reset() {
	s.modelCache.Reset()
}
