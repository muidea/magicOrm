package provider

import (
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
)

// Provider model provider
type Provider interface {
	RegisterModel(entity interface{}) (err error)

	UnregisterModel(entity interface{})

	GetEntityModel(entity interface{}) (ret model.Model, err error)

	GetValueModel(val reflect.Value) (ret model.Model, err error)

	GetSliceValueModel(val reflect.Value) (retModel model.Model, retVal reflect.Value, retErr error)

	GetTypeModel(vType model.Type) (ret model.Model, err error)

	GetValueStr(vType model.Type, vVal model.Value) (ret string, err error)

	GetModelDependValue(vModel model.Model, vVal model.Value) (ret []reflect.Value, err error)

	Reset()
}

// NewLocalProvider model provider
func NewLocalProvider() Provider {
	return local.New()
}

// NewRemoteProvider model provider
func NewRemoteProvider() Provider {
	return remote.New()
}
