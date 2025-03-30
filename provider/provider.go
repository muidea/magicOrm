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
	RegisterModel(entity any) (ret model.Model, err *cd.Error)

	UnregisterModel(entity any) (err *cd.Error)

	GetEntityType(entity any) (ret model.Type, err *cd.Error)

	GetEntityValue(entity any) (ret model.Value, err *cd.Error)

	GetEntityModel(entity any) (ret model.Model, err *cd.Error)

	GetEntityFilter(entity any, viewSpec model.ViewDeclare) (ret model.Filter, err *cd.Error)

	GetTypeModel(vType model.Type) (ret model.Model, err *cd.Error)

	GetModelFilter(vModel model.Model) (ret model.Filter, err *cd.Error)

	GetTypeFilter(vType model.Type, viewSpec model.ViewDeclare) (ret model.Filter, err *cd.Error)

	SetModelValue(vModel model.Model, vVal model.Value) (ret model.Model, err *cd.Error)

	EncodeValue(vVal any, vType model.Type) (ret any, err *cd.Error)

	DecodeValue(vVal any, vType model.Type) (ret any, err *cd.Error)

	Owner() string

	Reset()
}

// NewRemoteProvider model provider
func NewLocalProvider(owner string) Provider {
	ret := &providerImpl{
		owner:              owner,
		modelCache:         model.NewCache(),
		getEntityTypeFunc:  local.GetEntityType,
		getEntityValueFunc: local.GetEntityValue,
		getEntityModelFunc: local.GetEntityModel,
		getModelFilterFunc: local.GetModelFilter,
		setModelValueFunc:  local.SetModelValue,
		encodeValueFunc:    local.EncodeValue,
		decodeValueFunc:    local.DecodeValue,
	}

	return ret
}

// NewRemoteProvider model provider
func NewRemoteProvider(owner string) Provider {
	ret := &providerImpl{
		owner:              owner,
		modelCache:         model.NewCache(),
		getEntityTypeFunc:  remote.GetEntityType,
		getEntityValueFunc: remote.GetEntityValue,
		getEntityModelFunc: remote.GetEntityModel,
		getModelFilterFunc: remote.GetModelFilter,
		setModelValueFunc:  remote.SetModelValue,
		encodeValueFunc:    remote.EncodeValue,
		decodeValueFunc:    remote.DecodeValue,
	}

	return ret
}

type providerImpl struct {
	owner string

	modelCache model.Cache

	getEntityTypeFunc  func(any) (model.Type, *cd.Error)
	getEntityValueFunc func(any) (model.Value, *cd.Error)
	getEntityModelFunc func(any) (model.Model, *cd.Error)
	getModelFilterFunc func(model.Model) (model.Filter, *cd.Error)
	setModelValueFunc  func(model.Model, model.Value) (model.Model, *cd.Error)
	encodeValueFunc    func(any, model.Type) (any, *cd.Error)
	decodeValueFunc    func(any, model.Type) (any, *cd.Error)
}

func (s *providerImpl) RegisterModel(entity any) (ret model.Model, err *cd.Error) {
	entityModel, entityErr := s.getEntityModelFunc(entity)
	if entityErr != nil {
		err = entityErr
		log.Errorf("RegisterModel failed, s.getModelFunc error:%v", err.Error())
		return
	}

	pkgKey := entityModel.GetPkgKey()
	curModel := s.modelCache.Fetch(pkgKey)
	if curModel != nil {
		ret = entityModel
		return
	}

	// 这里主动Copy一份，避免污染原始的entity
	s.modelCache.Put(pkgKey, entityModel.Copy(model.MetaView))
	ret = entityModel
	return
}

func (s *providerImpl) UnregisterModel(entity any) (err *cd.Error) {
	modelType, modelErr := s.getEntityTypeFunc(entity)
	if modelErr != nil {
		err = modelErr
		log.Errorf("UnregisterModel failed, s.getTypeFunc error:%v", err.Error())
		return
	}

	modelType = modelType.Elem()
	curModel := s.modelCache.Fetch(modelType.GetPkgKey())
	if curModel != nil {
		s.modelCache.Remove(curModel.GetPkgKey())
	}
	return
}

func (s *providerImpl) GetEntityType(entity any) (ret model.Type, err *cd.Error) {
	ret, err = s.getEntityTypeFunc(entity)
	if err != nil {
		log.Errorf("GetEntityType failed, s.getTypeFunc error:%v", err.Error())
	}
	return
}

func (s *providerImpl) GetEntityValue(entity any) (ret model.Value, err *cd.Error) {
	ret, err = s.getEntityValueFunc(entity)
	if err != nil {
		log.Errorf("GetEntityValue failed, s.getValueFunc error:%v", err.Error())
	}
	return
}

func (s *providerImpl) GetEntityModel(entity any) (ret model.Model, err *cd.Error) {
	ret, err = s.checkEntityModel(entity, model.MetaView)
	if err != nil {
		log.Errorf("GetEntityModel failed, s.checkEntityModel error:%v", err.Error())
	}
	return
}

// checkEntityModel check entity model
// entity 可以是struct model type or model value
// 这里需要先进行判断
func (s *providerImpl) checkEntityModel(entity any, viewSpec model.ViewDeclare) (ret model.Model, err *cd.Error) {
	entityType, entityTypeErr := s.getEntityTypeFunc(entity)
	if entityTypeErr != nil {
		err = entityTypeErr
		log.Errorf("checkEntityModel failed, s.getEntityTypeFunc error:%v", err.Error())
		return
	}

	pkgKey := entityType.Elem().GetPkgKey()
	curModelVal := s.modelCache.Fetch(pkgKey)
	if curModelVal == nil {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("can't fetch model, PkgKey:%s", pkgKey))
		log.Errorf("checkEntityModel failed, error:%v", err.Error())
		return
	}
	entityValue, entityValueErr := s.getEntityValueFunc(entity)
	if entityValueErr != nil {
		// 到这里说明entity只是model type,不是model value
		ret = curModelVal.Copy(viewSpec)
		return
	}

	entityModelVal, entityModelErr := s.setModelValueFunc(curModelVal.Copy(viewSpec), entityValue)
	if entityModelErr != nil {
		err = entityModelErr
		log.Errorf("checkEntityModel failed, s.setModelValueFunc error:%v", err.Error())
	}

	ret = entityModelVal
	return
}

func (s *providerImpl) GetEntityFilter(entity any, viewSpec model.ViewDeclare) (ret model.Filter, err *cd.Error) {
	entityModelVal, entityModelErr := s.checkEntityModel(entity, viewSpec)
	if entityModelErr != nil {
		err = entityModelErr
		log.Errorf("GetEntityFilter failed, s.checkEntityModel error:%v", err.Error())
		return
	}

	ret, err = s.getModelFilterFunc(entityModelVal)
	if err != nil {
		log.Errorf("GetEntityFilter failed, s.getModelFilterFunc error:%v", err.Error())
	}
	return
}

func (s *providerImpl) GetTypeModel(vType model.Type) (ret model.Model, err *cd.Error) {
	if model.IsBasic(vType) {
		err = cd.NewError(cd.UnExpected, "illegal type value, type pkgKey:"+vType.GetPkgKey())
		log.Errorf("GetTypeModel failed, error:%v", err.Error())
		return
	}

	pkgKey := vType.Elem().GetPkgKey()
	typeModelVal := s.modelCache.Fetch(pkgKey)
	if typeModelVal == nil {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("can't fetch type model, must register type entity first, PkgKey:%s", pkgKey))
		log.Errorf("GetTypeModel failed, error:%v", err.Error())
		return
	}

	ret = typeModelVal.Copy(model.MetaView)
	return
}

func (s *providerImpl) GetModelFilter(vModel model.Model) (ret model.Filter, err *cd.Error) {
	if vModel == nil {
		err = cd.NewError(cd.UnExpected, "illegal model value")
		log.Errorf("GetModelFilter failed, error:%v", err.Error())
		return
	}

	pkgKey := vModel.GetPkgKey()
	curModelVal := s.modelCache.Fetch(pkgKey)
	if curModelVal == nil {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("can't fetch model, PkgKey:%s", pkgKey))
		log.Errorf("GetModelFilter failed, error:%v", err.Error())
		return
	}

	filterVal, filterErr := s.getModelFilterFunc(vModel)
	if filterErr != nil {
		err = filterErr
		log.Errorf("GetModelFilter failed, getFilterFunc error:%v", err.Error())
		return
	}

	ret = filterVal
	return
}

func (s *providerImpl) GetTypeFilter(vType model.Type, viewSpec model.ViewDeclare) (ret model.Filter, err *cd.Error) {
	if vType == nil {
		err = cd.NewError(cd.UnExpected, "illegal type value")
		log.Errorf("GetTypeFilter failed, error:%v", err.Error())
		return
	}

	pkgKey := vType.GetPkgKey()
	curModelVal := s.modelCache.Fetch(pkgKey)
	if curModelVal == nil {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("can't fetch model, PkgKey:%s", pkgKey))
		log.Errorf("GetTypeFilter failed, error:%v", err.Error())
		return
	}

	filterVal, filterErr := s.getModelFilterFunc(curModelVal.Copy(viewSpec))
	if filterErr != nil {
		err = filterErr
		log.Errorf("GetTypeFilter failed, getFilterFunc error:%v", err.Error())
		return
	}

	ret = filterVal
	return
}

func (s *providerImpl) SetModelValue(vModel model.Model, vVal model.Value) (ret model.Model, err *cd.Error) {
	pkgKey := vModel.GetPkgKey()
	curModel := s.modelCache.Fetch(pkgKey)
	if curModel == nil {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("can't fetch model, PkgKey:%s", pkgKey))
		log.Errorf("SetModelValue failed, error:%v", err.Error())
		return
	}

	ret, err = s.setModelValueFunc(vModel, vVal)
	if err != nil {
		log.Errorf("SetModelValue failed, s.setModelValueFunc error:%v", err.Error())
	}
	return
}

func (s *providerImpl) EncodeValue(vVal any, vType model.Type) (ret any, err *cd.Error) {
	if model.IsBasic(vType) {
		ret, err = s.encodeValueFunc(vVal, vType)
		if err != nil {
			log.Errorf("EncodeValue failed, s.encodeValueFunc error:%v", err.Error())
		}
		return
	}

	pkgKey := vType.Elem().GetPkgKey()
	curModelVal := s.modelCache.Fetch(pkgKey)
	if curModelVal == nil {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("can't fetch model, PkgKey:%s", pkgKey))
		log.Errorf("EncodeValue failed, error:%v", err.Error())
		return
	}

	eVal, eErr := s.getEntityValueFunc(vVal)
	if eErr != nil {
		err = eErr
		log.Errorf("EncodeValue failed, s.getEntityValueFunc error:%v", err.Error())
		return
	}
	vModelVal, vModelErr := s.setModelValueFunc(curModelVal.Copy(model.LiteView), eVal)
	if vModelErr != nil {
		err = vModelErr
		log.Errorf("EncodeValue failed, s.setModelValueFunc error:%v", err.Error())
		return
	}

	pkField := vModelVal.GetPrimaryField()
	ret, err = s.encodeValueFunc(pkField.GetValue().Get(), pkField.GetType())
	return
}

func (s *providerImpl) DecodeValue(vVal any, vType model.Type) (ret any, err *cd.Error) {
	ret, err = s.decodeValueFunc(vVal, vType)
	if err != nil {
		log.Errorf("DecodeValue failed, s.decodeValueFunc error:%v", err.Error())
	}
	return
}

//func (s *providerImpl) GetNewValue(valueDeclare model.ValueDeclare) (ret model.Value) {
//	ret = s.getNewValue(valueDeclare)
//	return
//}

func (s *providerImpl) Owner() string {
	return s.owner
}

func (s *providerImpl) Reset() {
	s.modelCache.Reset()
}
