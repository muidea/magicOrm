package provider

import (
	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
)

// Option defines a functional option for configuring providerImpl
type Option func(*providerImpl)

// WithModelCache sets a custom model cache
func WithModelCache(cache models.Cache) Option {
	return func(p *providerImpl) {
		p.modelCache = cache
	}
}

// WithValueValidator sets a custom value validator
func WithValueValidator(validator models.ValueValidator) Option {
	return func(p *providerImpl) {
		p.valueValidator = validator
	}
}

// WithEntityTypeFunc sets a custom entity type function
func WithEntityTypeFunc(fn func(any) (models.Type, *cd.Error)) Option {
	return func(p *providerImpl) {
		p.getEntityTypeFunc = fn
	}
}

// WithEntityValueFunc sets a custom entity value function
func WithEntityValueFunc(fn func(any) (models.Value, *cd.Error)) Option {
	return func(p *providerImpl) {
		p.getEntityValueFunc = fn
	}
}

// WithEntityModelFunc sets a custom entity model function
func WithEntityModelFunc(fn func(any, models.ValueValidator) (models.Model, *cd.Error)) Option {
	return func(p *providerImpl) {
		p.getEntityModelFunc = fn
	}
}

// WithModelFilterFunc sets a custom model filter function
func WithModelFilterFunc(fn func(models.Model) (models.Filter, *cd.Error)) Option {
	return func(p *providerImpl) {
		p.getModelFilterFunc = fn
	}
}

// WithSetModelValueFunc sets a custom set model value function
func WithSetModelValueFunc(fn func(models.Model, models.Value, bool) (models.Model, *cd.Error)) Option {
	return func(p *providerImpl) {
		p.setModelValueFunc = fn
	}
}

// WithEncodeValueFunc sets a custom encode value function
func WithEncodeValueFunc(fn func(any, models.Type) (any, *cd.Error)) Option {
	return func(p *providerImpl) {
		p.encodeValueFunc = fn
	}
}

// WithDecodeValueFunc sets a custom decode value function
func WithDecodeValueFunc(fn func(any, models.Type) (any, *cd.Error)) Option {
	return func(p *providerImpl) {
		p.decodeValueFunc = fn
	}
}

// newProvider creates a new provider with the given options
func newProvider(owner string, opts ...Option) *providerImpl {
	p := &providerImpl{
		owner:              owner,
		modelCache:         models.NewCache(),
		valueValidator:     nil,
		getEntityTypeFunc:  nil,
		getEntityValueFunc: nil,
		getEntityModelFunc: nil,
		getModelFilterFunc: nil,
		setModelValueFunc:  nil,
		encodeValueFunc:    nil,
		decodeValueFunc:    nil,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// NewLocalProviderWithOptions creates a local provider with functional options
func NewLocalProviderWithOptions(owner string, opts ...Option) Provider {
	p := newProvider(owner, opts...)

	// Set default local functions if not overridden by options
	if p.getEntityTypeFunc == nil {
		p.getEntityTypeFunc = local.GetEntityType
	}
	if p.getEntityValueFunc == nil {
		p.getEntityValueFunc = local.GetEntityValue
	}
	if p.getEntityModelFunc == nil {
		p.getEntityModelFunc = local.GetEntityModel
	}
	if p.getModelFilterFunc == nil {
		p.getModelFilterFunc = local.GetModelFilter
	}
	if p.setModelValueFunc == nil {
		p.setModelValueFunc = local.SetModelValue
	}
	if p.encodeValueFunc == nil {
		p.encodeValueFunc = local.EncodeValue
	}
	if p.decodeValueFunc == nil {
		p.decodeValueFunc = local.DecodeValue
	}

	return p
}

// NewRemoteProviderWithOptions creates a remote provider with functional options
func NewRemoteProviderWithOptions(owner string, opts ...Option) Provider {
	p := newProvider(owner, opts...)

	// Set default remote functions if not overridden by options
	if p.getEntityTypeFunc == nil {
		p.getEntityTypeFunc = remote.GetEntityType
	}
	if p.getEntityValueFunc == nil {
		p.getEntityValueFunc = remote.GetEntityValue
	}
	if p.getEntityModelFunc == nil {
		p.getEntityModelFunc = remote.GetEntityModel
	}
	if p.getModelFilterFunc == nil {
		p.getModelFilterFunc = remote.GetModelFilter
	}
	if p.setModelValueFunc == nil {
		p.setModelValueFunc = remote.SetModelValue
	}
	if p.encodeValueFunc == nil {
		p.encodeValueFunc = remote.EncodeValue
	}
	if p.decodeValueFunc == nil {
		p.decodeValueFunc = remote.DecodeValue
	}

	return p
}
