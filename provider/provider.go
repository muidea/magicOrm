package provider

import (
	"reflect"

	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/provider/local"
)

// Provider model provider
type Provider interface {
	GetObjectModel(obj interface{}) (ret model.Model, err error)

	GetValueModel(val reflect.Value) (ret model.Model, err error)

	GetTypeModel(vType model.Type) (ret model.Model, err error)

	GetValueStr(vType model.Type, vVal model.Value) (ret string, err error)

	GetModelDependValue(vModel model.Model, vVal model.Value) (ret []reflect.Value, err error)

	Reset()
}

// NewProvider model provider
func NewProvider() Provider {
	return local.New()
}
