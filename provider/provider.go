package provider

import (
	"fmt"

	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"

	"github.com/muidea/magicOrm/model"
)

type Provider interface {
	RegisterModel(entity interface{}) (ret model.Model, err error)

	UnregisterModel(entity interface{})

	GetEntityType(entity interface{}) (ret model.Type, err error)

	GetEntityValue(entity interface{}) (ret model.Value, err error)

	GetEntityModel(entity interface{}) (ret model.Model, err error)

	GetEntityFilter(entity interface{}) (ret model.Filter, err error)

	GetValueModel(vVal model.Value, vType model.Type) (ret model.Model, err error)

	GetTypeModel(vType model.Type) (ret model.Model, err error)

	GetTypeFilter(vType model.Type) (ret model.Filter, err error)

	EncodeValue(vVal model.Value, vType model.Type) (ret interface{}, err error)

	DecodeValue(vVal interface{}, vType model.Type) (ret model.Value, err error)

	ElemDependValue(val model.Value) (ret []model.Value, err error)

	AppendSliceValue(sliceVal model.Value, val model.Value) (ret model.Value, err error)

	IsAssigned(vVal model.Value, vType model.Type) bool

	GetValue(valueDeclare model.ValueDeclare) (ret model.Value)

	Owner() string

	Reset()
}

// NewLocalProvider model provider
func NewLocalProvider(owner string) Provider {
	ret := &providerImpl{
		owner:                owner,
		modelCache:           model.NewCache(),
		getTypeFunc:          local.GetEntityType,
		getValueFunc:         local.GetEntityValue,
		getModelFunc:         local.GetEntityModel,
		getFilterFunc:        local.GetModelFilter,
		setModelValueFunc:    local.SetModelValue,
		elemDependValueFunc:  local.ElemDependValue,
		appendSliceValueFunc: local.AppendSliceValue,
		encodeValueFunc:      local.EncodeValue,
		decodeValueFunc:      local.DecodeValue,
		getValue:             local.GetValue,
	}

	return ret
}

// NewRemoteProvider model provider
func NewRemoteProvider(owner string) Provider {
	ret := &providerImpl{
		owner:                owner,
		modelCache:           model.NewCache(),
		getTypeFunc:          remote.GetEntityType,
		getValueFunc:         remote.GetEntityValue,
		getModelFunc:         remote.GetEntityModel,
		getFilterFunc:        remote.GetModelFilter,
		setModelValueFunc:    remote.SetModelValue,
		elemDependValueFunc:  remote.ElemDependValue,
		appendSliceValueFunc: remote.AppendSliceValue,
		encodeValueFunc:      remote.EncodeValue,
		decodeValueFunc:      remote.DecodeValue,
		getValue:             remote.GetValue,
	}

	return ret
}

type providerImpl struct {
	owner string

	modelCache model.Cache

	getTypeFunc          func(interface{}) (model.Type, error)
	getValueFunc         func(interface{}) (model.Value, error)
	getModelFunc         func(interface{}) (model.Model, error)
	getFilterFunc        func(model.Model) (model.Filter, error)
	setModelValueFunc    func(model.Model, model.Value) (model.Model, error)
	elemDependValueFunc  func(model.Value) ([]model.Value, error)
	appendSliceValueFunc func(model.Value, model.Value) (model.Value, error)
	encodeValueFunc      func(model.Value, model.Type, model.Cache) (interface{}, error)
	decodeValueFunc      func(interface{}, model.Type, model.Cache) (model.Value, error)
	getValue             func(declare model.ValueDeclare) model.Value
}

func (s *providerImpl) RegisterModel(entity interface{}) (ret model.Model, err error) {
	modelType, modelErr := s.getTypeFunc(entity)
	if modelErr != nil {
		err = modelErr
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

func (s *providerImpl) UnregisterModel(entity interface{}) {
	modelType, modelErr := s.getTypeFunc(entity)
	if modelErr != nil {
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

func (s *providerImpl) GetEntityModel(entity interface{}) (ret model.Model, err error) {
	entityType, entityErr := s.getTypeFunc(entity)
	if entityErr != nil {
		err = fmt.Errorf("illegal entity, must be struct entity")
		return
	}

	// must check if register already
	entityModel := s.modelCache.Fetch(entityType.GetPkgKey())
	if entityModel == nil {
		err = fmt.Errorf("can't fetch entity model, must register entity first, entity PkgKey:%s", entityType.GetPkgKey())
		return
	}

	if model.IsSliceType(entityType.GetValue()) {
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

func (s *providerImpl) GetEntityFilter(entity interface{}) (ret model.Filter, err error) {
	vType, vErr := s.getTypeFunc(entity)
	if vErr != nil {
		err = vErr
		return
	}
	vType = vType.Elem()
	typeModel := s.modelCache.Fetch(vType.GetPkgKey())
	if typeModel == nil {
		err = fmt.Errorf("can't fetch type model, must register type entity first, PkgKey:%s", vType.GetPkgKey())
		return
	}

	ret, err = s.getFilterFunc(typeModel.Copy())
	return
}

func (s *providerImpl) GetValueModel(vVal model.Value, vType model.Type) (ret model.Model, err error) {
	typeModel := s.modelCache.Fetch(vType.GetPkgKey())
	if typeModel == nil {
		err = fmt.Errorf("can't fetch type model, must register type entity first")
		return
	}

	ret, err = s.setModelValueFunc(typeModel.Copy(), vVal)
	return
}

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

	ret = typeModel.Copy()
	return
}

func (s *providerImpl) GetTypeFilter(vType model.Type) (ret model.Filter, err error) {
	if vType.IsBasic() {
		return
	}
	vType = vType.Elem()
	typeModel := s.modelCache.Fetch(vType.GetPkgKey())
	if typeModel == nil {
		err = fmt.Errorf("can't fetch type filter, must register type entity first, PkgKey:%s", vType.GetPkgKey())
		return
	}

	ret, err = s.getFilterFunc(typeModel.Copy())
	return
}

func (s *providerImpl) EncodeValue(vVal model.Value, vType model.Type) (ret interface{}, err error) {
	ret, err = s.encodeValueFunc(vVal, vType, s.modelCache)
	return
}

func (s *providerImpl) DecodeValue(vVal interface{}, vType model.Type) (ret model.Value, err error) {
	ret, err = s.decodeValueFunc(vVal, vType, s.modelCache)
	return
}

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
	originVal := vType.Interface()
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

func (s *providerImpl) GetValue(valueDeclare model.ValueDeclare) (ret model.Value) {
	ret = s.getValue(valueDeclare)
	return
}

func (s *providerImpl) Owner() string {
	return s.owner
}

func (s *providerImpl) Reset() {
	s.modelCache.Reset()
}
