package provider

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/models"
	"log/slog"
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

type providerImpl struct {
	owner string

	modelCache models.Cache

	valueValidator models.ValueValidator

	getEntityTypeFunc  func(any) (models.Type, *cd.Error)
	getEntityValueFunc func(any) (models.Value, *cd.Error)
	getEntityModelFunc func(any, models.ValueValidator) (models.Model, *cd.Error)
	getModelFilterFunc func(models.Model) (models.Filter, *cd.Error)
	setModelValueFunc  func(models.Model, models.Value, bool) (models.Model, *cd.Error)
	encodeValueFunc    func(any, models.Type) (any, *cd.Error)
	decodeValueFunc    func(any, models.Type) (any, *cd.Error)
}

func (s *providerImpl) RegisterModel(entity any) (ret models.Model, err *cd.Error) {
	entityModel, entityErr := s.getEntityModelFunc(entity, s.valueValidator)
	if entityErr != nil {
		err = entityErr
		slog.Error("provider error", "method", "RegisterModel", "operation", "s.getModelFunc", "error", err.Error())
		return
	}

	pkgKey := entityModel.GetPkgKey()
	curModel := s.modelCache.Fetch(pkgKey)
	if curModel != nil {
		ret = entityModel
		return
	}

	// 这里主动Copy一份，避免污染原始的entity
	s.modelCache.Put(pkgKey, entityModel.Copy(models.MetaView))
	ret = entityModel
	return
}

func (s *providerImpl) UnregisterModel(entity any) (err *cd.Error) {
	modelType, modelErr := s.getEntityTypeFunc(entity)
	if modelErr != nil {
		err = modelErr
		slog.Error("provider error", "method", "UnregisterModel", "operation", "s.getTypeFunc", "error", err.Error())
		return
	}

	modelType = modelType.Elem()
	curModel := s.modelCache.Fetch(modelType.GetPkgKey())
	if curModel != nil {
		s.modelCache.Remove(curModel.GetPkgKey())
	}
	return
}

func (s *providerImpl) GetEntityType(entity any) (ret models.Type, err *cd.Error) {
	ret, err = s.getEntityTypeFunc(entity)
	if err != nil {
		slog.Error("provider error", "method", "GetEntityType", "operation", "s.getTypeFunc", "error", err.Error())
	}
	return
}

func (s *providerImpl) GetEntityValue(entity any) (ret models.Value, err *cd.Error) {
	ret, err = s.getEntityValueFunc(entity)
	if err != nil {
		slog.Error("provider error", "method", "GetEntityValue", "operation", "s.getValueFunc", "error", err.Error())
	}
	return
}

func (s *providerImpl) GetEntityModel(entity any, disableValidator bool) (ret models.Model, err *cd.Error) {
	ret, err = s.checkEntityModel(entity, models.MetaView, disableValidator)
	if err != nil {
		slog.Error("provider error", "method", "GetEntityModel", "operation", "s.checkEntityModel", "error", err.Error())
	}
	return
}

// checkEntityModel check entity model
// entity 可以是struct model type or model value
// 这里需要先进行判断
func (s *providerImpl) checkEntityModel(entity any, viewSpec models.ViewDeclare, disableValidator bool) (ret models.Model, err *cd.Error) {
	entityType, entityTypeErr := s.getEntityTypeFunc(entity)
	if entityTypeErr != nil {
		err = entityTypeErr
		slog.Error("provider error", "method", "checkEntityModel", "operation", "s.getEntityTypeFunc", "error", err.Error())
		return
	}

	pkgKey := entityType.Elem().GetPkgKey()
	curModelVal := s.modelCache.Fetch(pkgKey)
	if curModelVal == nil {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("can't fetch model, PkgKey:%s", pkgKey))
		slog.Error("provider error", "method", "checkEntityModel", "operation", "modelCache.Fetch", "error", err.Error())
		return
	}
	entityValue, entityValueErr := s.getEntityValueFunc(entity)
	if entityValueErr != nil {
		// 到这里说明entity只是model type,不是model value
		ret = curModelVal.Copy(viewSpec)
		return
	}

	entityModelVal, entityModelErr := s.setModelValueFunc(curModelVal.Copy(viewSpec), entityValue, disableValidator)
	if entityModelErr != nil {
		err = entityModelErr
		slog.Error("provider error", "method", "checkEntityModel", "operation", "s.setModelValueFunc", "error", err.Error())
	}

	ret = entityModelVal
	return
}

func (s *providerImpl) GetEntityFilter(entity any, viewSpec models.ViewDeclare) (ret models.Filter, err *cd.Error) {
	entityModelVal, entityModelErr := s.checkEntityModel(entity, viewSpec, true)
	if entityModelErr != nil {
		err = entityModelErr
		slog.Error("provider error", "method", "GetEntityFilter", "operation", "s.checkEntityModel", "error", err.Error())
		return
	}

	ret, err = s.getModelFilterFunc(entityModelVal)
	if err != nil {
		slog.Error("provider error", "method", "GetEntityFilter", "operation", "s.getModelFilterFunc", "error", err.Error())
	}
	return
}

func (s *providerImpl) GetTypeModel(vType models.Type) (ret models.Model, err *cd.Error) {
	if vType == nil {
		err = cd.NewError(cd.IllegalParam, "vType is nil")
		slog.Error("GetTypeModel: vType is nil")
		return
	}
	if models.IsBasic(vType) {
		err = cd.NewError(cd.Unexpected, "illegal type value, type pkgKey:"+vType.GetPkgKey())
		slog.Error("GetTypeModel: basic type not supported", "pkgKey", vType.GetPkgKey(), "error", err.Error())
		return
	}

	pkgKey := vType.Elem().GetPkgKey()
	typeModelVal := s.modelCache.Fetch(pkgKey)
	if typeModelVal == nil {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("can't fetch type model, must register type entity first, PkgKey:%s", pkgKey))
		slog.Error("GetTypeModel: model not found", "pkgKey", pkgKey, "error", err.Error())
		return
	}

	ret = typeModelVal.Copy(models.MetaView)
	return
}

func (s *providerImpl) GetModelFilter(vModel models.Model) (ret models.Filter, err *cd.Error) {
	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "vModel is nil")
		slog.Error("GetModelFilter: vModel is nil")
		return
	}

	pkgKey := vModel.GetPkgKey()
	curModelVal := s.modelCache.Fetch(pkgKey)
	if curModelVal == nil {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("can't fetch model, PkgKey:%s", pkgKey))
		slog.Error("GetModelFilter: model not found", "pkgKey", pkgKey, "error", err.Error())
		return
	}

	filterVal, filterErr := s.getModelFilterFunc(vModel)
	if filterErr != nil {
		err = filterErr
		slog.Error("GetModelFilter getModelFilterFunc failed", "pkgKey", pkgKey, "error", filterErr.Error())
		return
	}

	ret = filterVal
	return
}

func (s *providerImpl) GetTypeFilter(vType models.Type, viewSpec models.ViewDeclare) (ret models.Filter, err *cd.Error) {
	if vType == nil {
		err = cd.NewError(cd.IllegalParam, "vType is nil")
		slog.Error("GetTypeFilter: vType is nil")
		return
	}

	pkgKey := vType.GetPkgKey()
	curModelVal := s.modelCache.Fetch(pkgKey)
	if curModelVal == nil {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("can't fetch model, PkgKey:%s", pkgKey))
		slog.Error("GetTypeFilter: model not found", "pkgKey", pkgKey, "error", err.Error())
		return
	}

	filterVal, filterErr := s.getModelFilterFunc(curModelVal.Copy(viewSpec))
	if filterErr != nil {
		err = filterErr
		slog.Error("GetTypeFilter getModelFilterFunc failed", "pkgKey", pkgKey, "error", filterErr.Error())
		return
	}

	ret = filterVal
	return
}

func (s *providerImpl) SetModelValue(vModel models.Model, vVal models.Value) (ret models.Model, err *cd.Error) {
	if vModel == nil {
		err = cd.NewError(cd.IllegalParam, "vModel is nil")
		slog.Error("SetModelValue: vModel is nil")
		return
	}
	pkgKey := vModel.GetPkgKey()
	curModel := s.modelCache.Fetch(pkgKey)
	if curModel == nil {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("can't fetch model, PkgKey:%s", pkgKey))
		slog.Error("SetModelValue: model not found", "pkgKey", pkgKey, "error", err.Error())
		return
	}

	ret, err = s.setModelValueFunc(vModel, vVal, true)
	if err != nil {
		slog.Error("SetModelValue setModelValueFunc failed", "pkgKey", pkgKey, "error", err.Error())
	}
	return
}

func (s *providerImpl) EncodeValue(vVal any, vType models.Type) (ret any, err *cd.Error) {
	if vType == nil {
		err = cd.NewError(cd.IllegalParam, "vType is nil")
		slog.Error("EncodeValue: vType is nil")
		return
	}
	if models.IsBasic(vType) {
		ret, err = s.encodeValueFunc(vVal, vType)
		if err != nil {
			slog.Error("EncodeValue encodeValueFunc failed", "pkgKey", vType.GetPkgKey(), "error", err.Error())
		}
		return
	}

	pkgKey := vType.Elem().GetPkgKey()
	curModelVal := s.modelCache.Fetch(pkgKey)
	if curModelVal == nil {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("can't fetch model, PkgKey:%s", pkgKey))
		slog.Error("EncodeValue: model not found", "pkgKey", pkgKey, "error", err.Error())
		return
	}

	eVal, eErr := s.getEntityValueFunc(vVal)
	if eErr != nil {
		ret, err = s.encodeRelationPrimaryShorthand(vVal, curModelVal.Copy(models.MetaView))
		if err != nil {
			slog.Error("EncodeValue getEntityValueFunc failed", "pkgKey", pkgKey, "error", eErr.Error())
		}
		return
	}
	vModelVal, vModelErr := s.setModelValueFunc(curModelVal.Copy(models.LiteView), eVal, true)
	if vModelErr != nil {
		err = vModelErr
		slog.Error("EncodeValue setModelValueFunc failed", "pkgKey", pkgKey, "error", vModelErr.Error())
		return
	}

	pkField := vModelVal.GetPrimaryField()
	if pkField == nil {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("relation model missing primary key, PkgKey:%s", pkgKey))
		slog.Error("EncodeValue relation model missing primary key", "pkgKey", pkgKey, "error", err.Error())
		return
	}
	if !models.IsAssignedField(pkField) {
		err = cd.NewError(cd.IllegalParam, fmt.Sprintf("relation entity primary key is unassigned, PkgKey:%s", pkgKey))
		slog.Error("EncodeValue relation primary key is unassigned", "pkgKey", pkgKey, "error", err.Error())
		return
	}
	ret, err = s.encodeValueFunc(pkField.GetValue().Get(), pkField.GetType())
	return
}

func (s *providerImpl) encodeRelationPrimaryShorthand(vVal any, relationModel models.Model) (ret any, err *cd.Error) {
	if relationModel == nil {
		err = cd.NewError(cd.IllegalParam, "relation model is nil")
		return
	}

	pkField := relationModel.GetPrimaryField()
	if pkField == nil {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("relation model missing primary key, PkgKey:%s", relationModel.GetPkgKey()))
		return
	}
	if !models.IsBasic(pkField.GetType()) {
		err = cd.NewError(cd.IllegalParam, fmt.Sprintf("relation primary field must be basic, PkgKey:%s, field:%s", relationModel.GetPkgKey(), pkField.GetName()))
		return
	}

	pkValue, pkErr := pkField.GetType().Interface(vVal)
	if pkErr != nil {
		err = pkErr
		return
	}
	if pkValue == nil || !pkValue.IsValid() {
		err = cd.NewError(cd.IllegalParam, fmt.Sprintf("relation primary key shorthand is invalid, PkgKey:%s, field:%s", relationModel.GetPkgKey(), pkField.GetName()))
		return
	}

	ret, err = s.encodeValueFunc(pkValue.Get(), pkField.GetType())
	return
}

func (s *providerImpl) DecodeValue(vVal any, vType models.Type) (ret any, err *cd.Error) {
	if vType == nil {
		err = cd.NewError(cd.IllegalParam, "vType is nil")
		slog.Error("DecodeValue: vType is nil")
		return
	}
	ret, err = s.decodeValueFunc(vVal, vType)
	if err != nil {
		slog.Error("DecodeValue decodeValueFunc failed", "pkgKey", vType.GetPkgKey(), "error", err.Error())
	}
	return
}

//func (s *providerImpl) GetNewValue(valueDeclare models.ValueDeclare) (ret models.Value) {
//	ret = s.getNewValue(valueDeclare)
//	return
//}

func (s *providerImpl) Owner() string {
	return s.owner
}

func (s *providerImpl) Reset() {
	s.modelCache.Reset()
}
