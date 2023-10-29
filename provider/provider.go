package provider

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
)

type Provider interface {
	RegisterModel(entity interface{}) (ret model.Model, err *cd.Result)

	UnregisterModel(entity interface{}) (ret model.Model, err *cd.Result)

	GetEntityType(entity interface{}) (ret model.Type, err *cd.Result)

	GetEntityValue(entity interface{}) (ret model.Value, err *cd.Result)

	GetEntityModel(entity interface{}) (ret model.Model, err *cd.Result)

	GetEntityFilter(entity interface{}) (ret model.Filter, err *cd.Result)

	GetModelFilter(vModel model.Model) (ret model.Filter, err *cd.Result)

	GetValueModel(vVal model.Value, vType model.Type) (ret model.Model, err *cd.Result)

	GetTypeModel(vType model.Type) (ret model.Model, err *cd.Result)

	GetTypeFilter(vType model.Type) (ret model.Filter, err *cd.Result)

	EncodeValue(vVal model.Value, vType model.Type) (ret interface{}, err *cd.Result)

	DecodeValue(vVal interface{}, vType model.Type) (ret model.Value, err *cd.Result)

	ElemDependValue(val model.Value) (ret []model.Value, err *cd.Result)

	AppendSliceValue(sliceVal model.Value, val model.Value) (ret model.Value, err *cd.Result)

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

	getTypeFunc          func(interface{}) (model.Type, *cd.Result)
	getValueFunc         func(interface{}) (model.Value, *cd.Result)
	getModelFunc         func(interface{}) (model.Model, *cd.Result)
	getFilterFunc        func(model.Model) (model.Filter, *cd.Result)
	setModelValueFunc    func(model.Model, model.Value) (model.Model, *cd.Result)
	elemDependValueFunc  func(model.Value) ([]model.Value, *cd.Result)
	appendSliceValueFunc func(model.Value, model.Value) (model.Value, *cd.Result)
	encodeValueFunc      func(model.Value, model.Type, model.Cache) (interface{}, *cd.Result)
	decodeValueFunc      func(interface{}, model.Type, model.Cache) (model.Value, *cd.Result)
	getValue             func(declare model.ValueDeclare) model.Value
}

func (s *providerImpl) RegisterModel(entity interface{}) (ret model.Model, err *cd.Result) {
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

		err = cd.NewError(cd.UnExpected, fmt.Sprintf("confluct object model, name:%s,pkgKey:%s", modelType.GetName(), modelType.GetPkgKey()))
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

func (s *providerImpl) UnregisterModel(entity interface{}) (ret model.Model, err *cd.Result) {
	modelType, modelErr := s.getTypeFunc(entity)
	if modelErr != nil {
		err = modelErr
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
	ret = curModel
	return
}

func (s *providerImpl) GetEntityType(entity interface{}) (ret model.Type, err *cd.Result) {
	ret, err = s.getTypeFunc(entity)
	return
}

func (s *providerImpl) GetEntityValue(entity interface{}) (ret model.Value, err *cd.Result) {
	ret, err = s.getValueFunc(entity)
	return
}

func (s *providerImpl) GetEntityModel(entity interface{}) (ret model.Model, err *cd.Result) {
	entityType, entityErr := s.getTypeFunc(entity)
	if entityErr != nil {
		err = entityErr
		log.Errorf("GetEntityModel failed, s.getTypeFunc error:%v", err.Error())
		return
	}

	// must check if register already
	entityModel := s.modelCache.Fetch(entityType.GetPkgKey())
	if entityModel == nil {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("GetEntityModel failed, can't fetch entity model, must register entity first, entity PkgKey:%s", entityType.GetPkgKey()))
		log.Errorf("GetEntityModel failed, s.modelCache.Fetch error:%v", err.Error())
		return
	}

	if entityType.IsSlice() {
		ret = entityModel.Copy()
		return
	}

	entityValue, entityErr := s.getValueFunc(entity)
	if entityErr != nil {
		ret = entityModel.Copy()
		// 获取entity值失败，说明entity只是类型定义不是值
		// 这里要当成获取Model成功继续处理
		//err = entityErr
		return
	}

	ret, err = s.setModelValueFunc(entityModel.Copy(), entityValue)
	if err != nil {
		log.Errorf("GetEntityModel failed, setModelValueFunc error:%v", err.Error())
	}
	return
}

func (s *providerImpl) GetEntityFilter(entity interface{}) (ret model.Filter, err *cd.Result) {
	vModel, vErr := s.GetEntityModel(entity)
	if vErr != nil {
		err = vErr
		log.Errorf("GetEntityFilter failed, s.GetEntityModel error:%v", err.Error())
		return
	}

	ret, err = s.GetModelFilter(vModel)
	return
}

func (s *providerImpl) GetModelFilter(vModel model.Model) (ret model.Filter, err *cd.Result) {
	if vModel == nil {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal model value"))
		return
	}

	filterVal, filterErr := s.getFilterFunc(vModel)
	if filterErr != nil {
		err = filterErr
		log.Errorf("GetEntityFilter failed, getFilterFunc error:%v", err.Error())
		return
	}

	_ = filterVal.ValueMask(vModel.Interface(true, 0))
	ret = filterVal
	return
}

func (s *providerImpl) GetValueModel(vVal model.Value, vType model.Type) (ret model.Model, err *cd.Result) {
	typeModel := s.modelCache.Fetch(vType.GetPkgKey())
	if typeModel == nil {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("can't fetch type model, must register type:%s", vType.GetPkgKey()))
		log.Errorf("GetValueModel failed, s.modelCache.Fetch error:%v", err.Error())
		return
	}

	ret, err = s.setModelValueFunc(typeModel.Copy(), vVal)
	if err != nil {
		log.Errorf("GetValueModel failed, s.setModelValueFunc error:%v", err.Error())
		return
	}

	return
}

func (s *providerImpl) GetTypeModel(vType model.Type) (ret model.Model, err *cd.Result) {
	if vType.IsBasic() {
		return
	}

	vType = vType.Elem()
	typeModel := s.modelCache.Fetch(vType.GetPkgKey())
	if typeModel == nil {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("can't fetch type model, must register type entity first, PkgKey:%s", vType.GetPkgKey()))
		log.Errorf("GetTypeModel failed, error:%v", err.Error())
		return
	}

	ret = typeModel.Copy()
	return
}

func (s *providerImpl) GetTypeFilter(vType model.Type) (ret model.Filter, err *cd.Result) {
	if vType.IsBasic() {
		return
	}
	vType = vType.Elem()
	typeModel := s.modelCache.Fetch(vType.GetPkgKey())
	if typeModel == nil {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("can't fetch type filter, must register type entity first, PkgKey:%s", vType.GetPkgKey()))
		log.Errorf("GetTypeFilter failed, error:%v", err.Error())
		return
	}

	ret, err = s.getFilterFunc(typeModel.Copy())
	if err != nil {
		log.Errorf("GetTypeFilter failed, s.getFilterFunc error:%v", err.Error())
		return
	}

	return
}

func (s *providerImpl) EncodeValue(vVal model.Value, vType model.Type) (ret interface{}, err *cd.Result) {
	ret, err = s.encodeValueFunc(vVal, vType, s.modelCache)
	if err != nil {
		log.Errorf("EncodeValue failed, s.encodeValueFunc error:%v", err.Error())
		return
	}

	return
}

func (s *providerImpl) DecodeValue(vVal interface{}, vType model.Type) (ret model.Value, err *cd.Result) {
	ret, err = s.decodeValueFunc(vVal, vType, s.modelCache)
	if err != nil {
		log.Errorf("DecodeValue failed, s.decodeValueFunc error:%v", err.Error())
		return
	}

	return
}

func (s *providerImpl) ElemDependValue(val model.Value) (ret []model.Value, err *cd.Result) {
	ret, err = s.elemDependValueFunc(val)
	if err != nil {
		log.Errorf("ElemDependValue failed, s.elemDependValueFunc error:%v", err.Error())
		return
	}

	return
}

func (s *providerImpl) AppendSliceValue(sliceVal model.Value, val model.Value) (ret model.Value, err *cd.Result) {
	ret, err = s.appendSliceValueFunc(sliceVal, val)
	if err != nil {
		log.Errorf("AppendSliceValue failed, s.appendSliceValueFunc error:%v", err.Error())
		return
	}

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
