package provider

import (
	"reflect"

	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/provider/local"
)

// Provider model provider
type Provider interface {
	GetObjectModel(objPtr interface{}) (ret model.Model, err error)

	GetTypeModel(modelType reflect.Type) (ret model.Model, err error)

	GetValueModel(modelVal reflect.Value) (ret model.Model, err error)

	GetValueStr(val reflect.Value) (ret string, err error)
}

// New model provider
func New() Provider {
	cache := model.NewCache()
	return local.New(cache)
}
