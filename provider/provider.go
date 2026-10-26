package provider

import (
	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/models"
)

type Provider interface {
	RegisterModel(entity any) (ret models.Model, err *cd.Error)

	UnregisterModel(entity any) (err *cd.Error)

	GetEntityType(entity any) (ret models.Type, err *cd.Error)

	GetEntityValue(entity any) (ret models.Value, err *cd.Error)

	GetEntityModel(entity any, disableValidator bool) (ret models.Model, err *cd.Error)

	GetEntityFilter(entity any, viewSpec models.ViewDeclare) (ret models.Filter, err *cd.Error)

	GetTypeModel(vType models.Type) (ret models.Model, err *cd.Error)

	GetModelFilter(vModel models.Model) (ret models.Filter, err *cd.Error)

	GetTypeFilter(vType models.Type, viewSpec models.ViewDeclare) (ret models.Filter, err *cd.Error)

	SetModelValue(vModel models.Model, vVal models.Value) (ret models.Model, err *cd.Error)

	EncodeValue(vVal any, vType models.Type) (ret any, err *cd.Error)

	DecodeValue(vVal any, vType models.Type) (ret any, err *cd.Error)

	Owner() string

	Reset()
}

// NewLocalProvider creates a local provider with the given owner and validator
// Deprecated: Use NewLocalProviderWithOptions for more flexible configuration
func NewLocalProvider(owner string, validator models.ValueValidator) Provider {
	return NewLocalProviderWithOptions(owner, WithValueValidator(validator))
}

// NewRemoteProvider creates a remote provider with the given owner and validator
// Deprecated: Use NewRemoteProviderWithOptions for more flexible configuration
func NewRemoteProvider(owner string, validator models.ValueValidator) Provider {
	return NewRemoteProviderWithOptions(owner, WithValueValidator(validator))
}
