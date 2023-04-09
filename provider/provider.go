package provider

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
	"github.com/muidea/magicOrm/util"
)

// Provider model provider
type Provider interface {
	RegisterModel(entity interface{}) (ret model.Model, err error)

	UnregisterModel(entity interface{})

	GetEntityType(entity interface{}) (ret model.Type, err error)

	GetEntityValue(entity interface{}) (ret model.Value, err error)

	GetEntityModel(entity interface{}) (ret model.Model, err error)

	GetValueModel(vVal model.Value, vType model.Type) (ret model.Model, err error)

	GetTypeModel(vType model.Type) (ret model.Model, err error)

	EncodeValue(vVal model.Value, vType model.Type) (ret interface{}, err error)

	DecodeValue(vVal interface{}, vType model.Type) (ret model.Value, err error)

	ElemDependValue(val model.Value) (ret []model.Value, err error)

	AppendSliceValue(sliceVal model.Value, val model.Value) (ret model.Value, err error)

	IsAssigned(vVal model.Value, vType model.Type) bool

	Owner() string

	Prefix() string

	Reset()
}

type providerImpl struct {
	owner  string
	prefix string

	modelCache model.Cache

	getTypeFunc          func(interface{}) (model.Type, error)
	getValueFunc         func(interface{}) (model.Value, error)
	getModelFunc         func(interface{}) (model.Model, error)
	setModelValueFunc    func(model.Model, model.Value) (model.Model, error)
	elemDependValueFunc  func(model.Value) ([]model.Value, error)
	appendSliceValueFunc func(model.Value, model.Value) (model.Value, error)
	encodeValueFunc      func(model.Value, model.Type, model.Cache) (interface{}, error)
	decodeValueFunc      func(interface{}, model.Type, model.Cache) (model.Value, error)
}

// RegisterModel RegisterObjectModel
func (s *providerImpl) RegisterModel(entity interface{}) (ret model.Model, err error) {
	modelType, modelErr := s.getTypeFunc(entity)
	if modelErr != nil {
		err = modelErr
		return
	}
	if modelType.IsBasic() {
		err = fmt.Errorf("illegal entity model, name:%s", modelType.GetName())
		return
	}

	modelType = modelType.Elem()
	curModel := s.modelCache.Fetch(modelType.GetPkgKey())
	if curModel != nil {
		if curModel.GetPkgPath() == modelType.GetPkgPath() {
			ret = curModel
			return
		}

		err = fmt.Errorf("confluct object model, name:%s,pkgKey:%s", modelType.GetName(), modelType.GetPkgKey())
		return
	}

	entityModel, entityErr := s.getModelFunc(entity)
	if entityErr != nil {
		err = entityErr
		return
	}

	s.modelCache.Put(entityModel.GetPkgKey(), entityModel)
	ret = entityModel
	return
}

// UnregisterModel register model
func (s *providerImpl) UnregisterModel(entity interface{}) {
	modelType, modelErr := s.getTypeFunc(entity)
	if modelErr != nil {
		return
	}
	if modelType.IsBasic() {
		return
	}

	modelType = modelType.Elem()
	curModel := s.modelCache.Fetch(modelType.GetPkgKey())
	if curModel != nil {
		if curModel.GetPkgPath() != modelType.GetPkgPath() {
			return
		}

		s.modelCache.Remove(curModel.GetPkgKey())
	}
	return
}

func (s *providerImpl) GetEntityType(entity interface{}) (ret model.Type, err error) {
	ret, err = s.getTypeFunc(entity)
	return
}

func (s *providerImpl) GetEntityValue(entity interface{}) (ret model.Value, err error) {
	ret, err = s.getValueFunc(entity)
	return
}

// GetEntityModel GetEntityModel
func (s *providerImpl) GetEntityModel(entity interface{}) (ret model.Model, err error) {
	entityType, entityErr := s.getTypeFunc(entity)
	if entityErr != nil || entityType.IsBasic() {
		err = fmt.Errorf("illegal entity, must be struct entity")
		return
	}

	// must check if register already
	entityModel := s.modelCache.Fetch(entityType.GetPkgKey())
	if entityModel == nil {
		err = fmt.Errorf("can't fetch entity model, must register entity first, entity PkgKey:%s", entityType.GetPkgKey())
		return
	}

	if entityModel.GetPkgPath() != entityType.GetPkgPath() {
		err = fmt.Errorf("illegal object entity, must register entity first")
		return
	}

	if util.IsSliceType(entityType.GetValue()) {
		ret = entityModel.Copy()
		return
	}

	entityValue, entityErr := s.getValueFunc(entity)
	if entityErr != nil {
		err = entityErr
		return
	}

	ret, err = s.setModelValueFunc(entityModel.Copy(), entityValue)
	return
}

// GetValueModel GetValueModel
func (s *providerImpl) GetValueModel(vVal model.Value, vType model.Type) (ret model.Model, err error) {
	typeModel := s.modelCache.Fetch(vType.GetPkgKey())
	if typeModel == nil {
		err = fmt.Errorf("can't fetch type model, must register type entity first")
		return
	}
	if typeModel.GetPkgPath() != vType.GetPkgPath() {
		err = fmt.Errorf("illegal object entity, must register entity first")
		return
	}

	ret, err = s.setModelValueFunc(typeModel.Copy(), vVal)
	return
}

// GetTypeModel GetTypeModel
func (s *providerImpl) GetTypeModel(vType model.Type) (ret model.Model, err error) {
	if vType.IsBasic() {
		return
	}
	vType = vType.Elem()
	typeModel := s.modelCache.Fetch(vType.GetPkgKey())
	if typeModel == nil {
		err = fmt.Errorf("can't fetch type model, must register type entity first, PkgKey:%s", vType.GetPkgKey())
		return
	}
	if typeModel.GetPkgPath() != vType.GetPkgPath() {
		err = fmt.Errorf("illegal object entity, must register entity first")
		return
	}

	ret = typeModel
	return
}

// GetValueStr GetValueStr
func (s *providerImpl) EncodeValue(vVal model.Value, vType model.Type) (ret interface{}, err error) {
	ret, err = s.encodeValueFunc(vVal, vType, s.modelCache)
	return
}

func (s *providerImpl) DecodeValue(vVal interface{}, vType model.Type) (ret model.Value, err error) {
	ret, err = s.decodeValueFunc(vVal, vType, s.modelCache)
	return
}

// GetValueDepend GetEntityValue depend values
func (s *providerImpl) ElemDependValue(val model.Value) (ret []model.Value, err error) {
	ret, err = s.elemDependValueFunc(val)
	return
}

func (s *providerImpl) AppendSliceValue(sliceVal model.Value, val model.Value) (ret model.Value, err error) {
	ret, err = s.appendSliceValueFunc(sliceVal, val)
	return
}

func (s *providerImpl) IsAssigned(vVal model.Value, vType model.Type) (ret bool) {
	if vVal.IsNil() {
		ret = false
		return
	}

	curVal := vVal
	originVal, _ := vType.Interface()
	curStr, curErr := s.encodeValueFunc(curVal, vType, s.modelCache)
	if curErr != nil {
		ret = false
		return
	}
	originStr, originErr := s.encodeValueFunc(originVal, vType, s.modelCache)
	if originErr != nil {
		ret = false
		return
	}

	ret = curStr != originStr
	return
}

func (s *providerImpl) Owner() string {
	return s.owner
}

func (s *providerImpl) Prefix() string {
	return s.prefix
}

// Reset Reset
func (s *providerImpl) Reset() {
	s.modelCache.Reset()
}

// NewLocalProvider model provider
func NewLocalProvider(owner, prefix string) Provider {
	ret := &providerImpl{
		owner:                owner,
		prefix:               prefix,
		modelCache:           model.NewCache(),
		getTypeFunc:          local.GetEntityType,
		getValueFunc:         local.GetEntityValue,
		getModelFunc:         local.GetEntityModel,
		setModelValueFunc:    local.SetModelValue,
		elemDependValueFunc:  local.ElemDependValue,
		appendSliceValueFunc: local.AppendSliceValue,
		encodeValueFunc:      local.EncodeValue,
		decodeValueFunc:      local.DecodeValue,
	}

	return ret
}

// NewRemoteProvider model provider
func NewRemoteProvider(owner, prefix string) Provider {
	ret := &providerImpl{
		owner:                owner,
		prefix:               prefix,
		modelCache:           model.NewCache(),
		getTypeFunc:          remote.GetEntityType,
		getValueFunc:         remote.GetEntityValue,
		getModelFunc:         remote.GetEntityModel,
		setModelValueFunc:    remote.SetModelValue,
		elemDependValueFunc:  remote.ElemDependValue,
		appendSliceValueFunc: remote.AppendSliceValue,
		encodeValueFunc:      remote.EncodeValue,
		decodeValueFunc:      remote.DecodeValue,
	}

	return ret
}
