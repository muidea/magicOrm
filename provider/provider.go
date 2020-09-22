package provider

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
	"github.com/muidea/magicOrm/util"
)

// Provider model provider
type Provider interface {
	RegisterModel(entity interface{}) (ret model.Model, err error)

	UnregisterModel(entity interface{})

	GetEntityType(entity interface{}) (ret model.Type, err error)

	GetEntityModel(entity interface{}) (ret model.Model, err error)

	GetValueModel(val reflect.Value) (ret model.Model, err error)

	GetTypeModel(vType model.Type) (ret model.Model, err error)

	GetValueStr(vType model.Type, vVal model.Value) (ret string, err error)

	ElemDependValue(val reflect.Value) (ret []reflect.Value, err error)

	AppendSliceValue(sliceVal reflect.Value, val reflect.Value) (ret reflect.Value, err error)

	Owner() string

	Reset()
}

type providerImpl struct {
	owner string

	localProvider bool
	modelCache    model.Cache

	getTypeFunc          func(reflect.Value) (model.Type, error)
	getModelFunc         func(reflect.Value) (model.Model, error)
	setModelValueFunc    func(model.Model, reflect.Value) (model.Model, error)
	elemDependValueFunc  func(model.Type, reflect.Value) ([]reflect.Value, error)
	appendSliceValueFunc func(reflect.Value, reflect.Value) (reflect.Value, error)
}

// RegisterModel RegisterObjectModel
func (s *providerImpl) RegisterModel(entity interface{}) (ret model.Model, err error) {
	entityValue := reflect.ValueOf(entity)
	modelType, modelErr := s.getTypeFunc(entityValue)
	if modelErr != nil {
		err = modelErr
		return
	}
	modelType = modelType.Depend()
	if modelType == nil {
		err = fmt.Errorf("illegal entity, must be a struct or slice struct")
		return
	}

	curModel := s.modelCache.Fetch(modelType.GetName())
	if curModel != nil {
		if curModel.GetPkgPath() == modelType.GetPkgPath() {
			ret = curModel
			return
		}

		err = fmt.Errorf("confluct object model, name:%s,pkgPath:%s", modelType.GetName(), modelType.GetPkgPath())
		return
	}

	entityModel, entityErr := s.getModelFunc(entityValue)
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
	entityValue := reflect.ValueOf(entity)
	modelType, modelErr := s.getTypeFunc(entityValue)
	if modelErr != nil {
		return
	}
	modelType = modelType.Depend()
	if modelType == nil {
		return
	}

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
	entityVal := reflect.ValueOf(entity)
	if util.IsNil(entityVal) {
		err = fmt.Errorf("illegal entity, nil value point")
		return
	}
	entityType, entityErr := s.getTypeFunc(entityVal)
	if entityErr != nil {
		err = entityErr
		return
	}

	ret = entityType
	return
}

// GetEntityModel GetEntityModel
func (s *providerImpl) GetEntityModel(entity interface{}) (ret model.Model, err error) {
	entityVal := reflect.ValueOf(entity)
	if entityVal.Kind() != reflect.Ptr {
		err = fmt.Errorf("illegal entity, must be value ptr")
		return
	}
	if util.IsNil(entityVal) {
		err = fmt.Errorf("illegal entity, nil value point")
		return
	}
	entityVal = reflect.Indirect(entityVal)
	entityType, entityErr := s.getTypeFunc(entityVal)
	if entityErr != nil || !util.IsStructType(entityType.GetValue()) {
		err = fmt.Errorf("illegal entity, must be struct value")
		return
	}
	if !entityVal.CanSet() {
		err = fmt.Errorf("illegal entity value, read only value")
		return
	}

	// must check if register already
	entityModel := s.modelCache.Fetch(entityType.GetName())
	if entityModel == nil {
		err = fmt.Errorf("can't fetch entity model, must register entity first")
		return
	}

	if entityModel.GetPkgPath() != entityType.GetPkgPath() {
		err = fmt.Errorf("illegal object entity, must register entity first")
		return
	}

	ret, err = s.setModelValueFunc(entityModel.Copy(), entityVal)
	return
}

// GetValueModel GetValueModel
func (s *providerImpl) GetValueModel(vVal reflect.Value) (ret model.Model, err error) {
	vType, vErr := s.getTypeFunc(vVal)
	if vErr != nil {
		err = vErr
		return
	}

	vType = vType.Depend()
	if util.IsBasicType(vType.GetValue()) {
		return
	}

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
	vType = vType.Depend()
	if vType == nil {
		return
	}
	if !util.IsStructType(vType.GetValue()) {
		return
	}

	typeModel := s.modelCache.Fetch(vType.GetName())
	if typeModel == nil {
		err = fmt.Errorf("can't fetch type model, must register type entity first")
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
func (s *providerImpl) GetValueStr(vType model.Type, vVal model.Value) (ret string, err error) {
	if vVal.IsNil() {
		return
	}

	if util.IsBasicType(vType.GetValue()) {
		ret, err = getBasicValue(vType, vVal.Get())
		return
	}

	if util.IsStructType(vType.GetValue()) {
		ret, err = s.getStructValue(vType, vVal)
		return
	}

	dType := vType.Depend()
	if util.IsBasicType(dType.GetValue()) {
		ret, err = helper.EncodeSliceValue(vVal.Get())
		return
	}

	ret, err = s.getSliceStructValue(vType, vVal)
	return
}

// GetValueDepend GetValue depend values
func (s *providerImpl) ElemDependValue(val reflect.Value) (ret []reflect.Value, err error) {
	vType, vErr := s.getTypeFunc(val)
	if vErr != nil {
		err = vErr
		return
	}
	if vType.Depend() == nil {
		return
	}

	vDependType := vType.Depend()
	typeModel := s.modelCache.Fetch(vDependType.GetName())
	if typeModel == nil {
		err = fmt.Errorf("can't fetch type model, must register type entity first")
		return
	}
	if typeModel.GetPkgPath() != vDependType.GetPkgPath() {
		err = fmt.Errorf("illegal object entity, must register entity first")
		return
	}

	val = reflect.Indirect(val)
	ret, err = s.elemDependValueFunc(vType, val)
	return
}

func (s *providerImpl) AppendSliceValue(sliceVal reflect.Value, val reflect.Value) (ret reflect.Value, err error) {
	return s.appendSliceValueFunc(sliceVal, val)
}

// Owner owner
func (s *providerImpl) Owner() string {
	return s.owner
}

// Reset Reset
func (s *providerImpl) Reset() {
	s.modelCache.Reset()
}

func (s *providerImpl) getStructValue(vType model.Type, vVal model.Value) (ret string, err error) {
	typeModel, typeErr := s.GetValueModel(vVal.Get())
	if typeErr != nil {
		err = typeErr
		return
	}

	pkField := typeModel.GetPrimaryField()
	return getBasicValue(pkField.GetType(), pkField.GetValue().Get())
}

func (s *providerImpl) getSliceStructValue(vType model.Type, vVal model.Value) (ret string, err error) {
	typeModel, typeErr := s.GetTypeModel(vType.Depend())
	if typeErr != nil {
		err = typeErr
		return
	}

	val := reflect.Indirect(vVal.Get())
	sliceVal, sliceErr := s.elemDependValueFunc(vType, val)
	if sliceErr != nil {
		err = sliceErr
		return
	}

	var sliceStrVal []string
	for idx := 0; idx < len(sliceVal); idx++ {
		vModel, vErr := s.setModelValueFunc(typeModel, sliceVal[idx])
		if vErr != nil {
			err = vErr
			return
		}

		pkField := vModel.GetPrimaryField()
		strVal, strErr := getBasicValue(pkField.GetType(), pkField.GetValue().Get())
		if strErr != nil {
			err = strErr
			log.Printf("getStructValue failed, err:%s", err.Error())
			return
		}

		sliceStrVal = append(sliceStrVal, strVal)
	}

	ret = strings.Join(sliceStrVal, ",")
	return
}

func getBasicValue(vType model.Type, val reflect.Value) (ret string, err error) {
	if util.IsNil(val) {
		return
	}

	switch vType.GetValue() {
	case util.TypeBooleanField:
		ret, err = helper.EncodeBoolValue(val)
	case util.TypeBitField, util.TypeSmallIntegerField, util.TypeInteger32Field, util.TypeBigIntegerField, util.TypeIntegerField:
		ret, err = helper.EncodeIntValue(val)
	case util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveInteger32Field, util.TypePositiveBigIntegerField, util.TypePositiveIntegerField:
		ret, err = helper.EncodeUintValue(val)
	case util.TypeFloatField, util.TypeDoubleField:
		ret, err = helper.EncodeFloatValue(val)
	case util.TypeStringField:
		strRet, strErr := helper.EncodeStringValue(val)
		if strErr != nil {
			err = strErr
			return
		}
		ret = fmt.Sprintf("'%s'", strRet)
	case util.TypeDateTimeField:
		strRet, strErr := helper.EncodeDateTimeValue(val)
		if strErr != nil {
			err = strErr
			return
		}

		ret = fmt.Sprintf("'%s'", strRet)
	default:
		err = fmt.Errorf("illegal value kind, type name:%v", vType.GetName())
	}

	return
}

// NewLocalProvider model provider
func NewLocalProvider(owner string) Provider {
	return &providerImpl{
		owner:                owner,
		localProvider:        true,
		modelCache:           model.NewCache(),
		getTypeFunc:          local.GetType,
		getModelFunc:         local.GetModel,
		setModelValueFunc:    local.SetModel,
		elemDependValueFunc:  local.ElemDependValue,
		appendSliceValueFunc: local.AppendSliceValue,
	}
}

// NewRemoteProvider model provider
func NewRemoteProvider(owner string) Provider {
	return &providerImpl{
		owner:                owner,
		localProvider:        false,
		modelCache:           model.NewCache(),
		getTypeFunc:          remote.GetType,
		getModelFunc:         remote.GetModel,
		setModelValueFunc:    remote.SetModel,
		elemDependValueFunc:  remote.ElemDependValue,
		appendSliceValueFunc: remote.AppendSliceValue,
	}
}
