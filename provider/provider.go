package provider

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
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

	GetValueStr(vVal model.Value, vType model.Type) (ret string, err error)

	ElemDependValue(val model.Value) (ret []model.Value, err error)

	AppendSliceValue(sliceVal model.Value, val model.Value) (ret model.Value, err error)

	IsAssigned(vVal model.Value, vType model.Type) bool

	Owner() string

	Reset()
}

type providerImpl struct {
	owner string

	modelCache model.Cache
	helper     helper.Helper

	getTypeFunc          func(interface{}) (model.Type, error)
	getValueFunc         func(interface{}) (model.Value, error)
	getModelFunc         func(interface{}) (model.Model, error)
	setModelValueFunc    func(model.Model, model.Value) (model.Model, error)
	elemDependValueFunc  func(model.Value) ([]model.Value, error)
	appendSliceValueFunc func(model.Value, model.Value) (model.Value, error)
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
	curModel := s.modelCache.Fetch(modelType.GetName())
	if curModel != nil {
		if curModel.GetPkgPath() == modelType.GetPkgPath() {
			ret = curModel
			return
		}

		err = fmt.Errorf("confluct object model, name:%s,pkgPath:%s", modelType.GetName(), modelType.GetPkgPath())
		return
	}

	entityModel, entityErr := s.getModelFunc(entity)
	if entityErr != nil {
		err = entityErr
		return
	}

	s.modelCache.Put(entityModel.GetName(), entityModel)
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
	curModel := s.modelCache.Fetch(modelType.GetName())
	if curModel != nil {
		if curModel.GetPkgPath() != modelType.GetPkgPath() {
			return
		}

		s.modelCache.Remove(curModel.GetName())
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
	entityModel := s.modelCache.Fetch(entityType.GetName())
	if entityModel == nil {
		err = fmt.Errorf("can't fetch entity model, must register entity first, entity Name:%s", entityType.GetName())
		return
	}

	if entityModel.GetPkgPath() != entityType.GetPkgPath() {
		err = fmt.Errorf("illegal object entity, must register entity first")
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
	typeModel := s.modelCache.Fetch(vType.GetName())
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
	typeModel := s.modelCache.Fetch(vType.GetName())
	if typeModel == nil {
		err = fmt.Errorf("can't fetch type model, must register type entity first, name:%s", vType.GetName())
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
func (s *providerImpl) GetValueStr(vVal model.Value, vType model.Type) (ret string, err error) {
	ret, err = s.helper.Encode(vVal, vType)
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
	originVal, _ := vType.Interface(nil)
	curStr, curErr := s.helper.Encode(curVal, vType)
	if curErr != nil {
		ret = false
		return
	}
	originStr, originErr := s.helper.Encode(originVal, vType)
	if originErr != nil {
		ret = false
	}

	ret = curStr != originStr
	return
}

// Owner owner
func (s *providerImpl) Owner() string {
	return s.owner
}

// Reset Reset
func (s *providerImpl) Reset() {
	s.modelCache.Reset()
}

// NewLocalProvider model provider
func NewLocalProvider(owner string) Provider {
	ret := &providerImpl{
		owner:                owner,
		modelCache:           model.NewCache(),
		getTypeFunc:          local.GetEntityType,
		getValueFunc:         local.GetEntityValue,
		getModelFunc:         local.GetEntityModel,
		setModelValueFunc:    local.SetModelValue,
		elemDependValueFunc:  local.ElemDependValue,
		appendSliceValueFunc: local.AppendSliceValue,
	}

	ret.helper = helper.New(ret.GetEntityValue, ret.GetValueModel)
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
		setModelValueFunc:    remote.SetModelValue,
		elemDependValueFunc:  remote.ElemDependValue,
		appendSliceValueFunc: remote.AppendSliceValue,
	}

	ret.helper = helper.New(ret.GetEntityValue, ret.GetValueModel)

	return ret
}
