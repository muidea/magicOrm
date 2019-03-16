package remote

import (
	"log"
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
	modelImpl, modelErr := GetObject(obj, s.modelCache)
	if modelErr != nil {
		err = modelErr
		log.Printf("getValueModel failed, err:%s", err.Error())
		return
	}

	ret = modelImpl
	return
}

// GetValueModel GetValueModel
func (s *Provider) GetValueModel(val reflect.Value) (ret model.Model, err error) {
	return
}

// GetTypeModel GetTypeModel
func (s *Provider) GetTypeModel(vType model.Type) (ret model.Model, err error) {
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
