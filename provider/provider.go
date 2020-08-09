package provider

import (
	"fmt"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
	"github.com/muidea/magicOrm/util"
	"log"
	"reflect"
	"strings"

	"github.com/muidea/magicOrm/model"
)

// Provider model provider
type Provider interface {
	RegisterModel(entity interface{}) (ret model.Model, err error)

	UnregisterModel(entity interface{})

	GetEntityModel(entity interface{}) (ret model.Model, err error)

	GetValueModel(val reflect.Value) (ret model.Model, err error)

	GetSliceValueModel(val reflect.Value) (ret model.Model, err error)

	GetTypeModel(vType model.Type) (ret model.Model, err error)

	GetValueStr(vType model.Type, vVal model.Value) (ret string, err error)

	GetDependValue(vVal model.Value) (ret []reflect.Value, err error)

	Owner() string

	Reset()
}

type providerImpl struct {
	owner string

	localProvider bool
	modelCache    model.Cache

	getTypeFunc       func(p reflect.Type) (ret model.Type, err error)
	getValueModelFunc func(p reflect.Value) (ret model.Model, err error)
	getTypeModelFunc  func(p reflect.Type) (ret model.Model, err error)
}

// RegisterModel RegisterObjectModel
func (s *providerImpl) RegisterModel(entity interface{}) (ret model.Model, err error) {
	entityType := reflect.TypeOf(entity)
	modelType, modelErr := s.getTypeFunc(entityType)
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
	if curModel.GetPkgPath() == modelType.GetPkgPath() {
		ret = curModel
		return
	}

	entityModel, entityErr := s.getTypeModelFunc(entityType)
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
	entityType := reflect.TypeOf(entity)
	modelType, modelErr := s.getTypeFunc(entityType)
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

// GetEntityModel GetEntityModel
func (s *providerImpl) GetEntityModel(objPtr interface{}) (ret model.Model, err error) {
	objVal := reflect.ValueOf(objPtr)
	if objVal.Kind() != reflect.Ptr {
		err = fmt.Errorf("illegal entity, must be value ptr")
		return
	}
	if util.IsNil(objVal) {
		err = fmt.Errorf("illegal entity, nil value point")
		return
	}
	objVal = reflect.Indirect(objVal)
	objType, objErr := s.getTypeFunc(objVal.Type())
	if objErr != nil || !util.IsStructType(objType.GetValue()) {
		err = fmt.Errorf("illegal entity, must be struct value")
		return
	}
	if !objVal.CanSet() {
		err = fmt.Errorf("illegal entity value, read only value")
		return
	}

	objModel, objErr := s.getValueModelFunc(objVal)
	if objErr != nil {
		err = objErr
		return
	}

	// must check if register already
	curModel := s.modelCache.Fetch(objModel.GetName())
	if curModel == nil {
		err = fmt.Errorf("can't fetch entity model, must register entity first")
		return
	}

	if curModel.GetPkgPath() != objModel.GetPkgPath() {
		err = fmt.Errorf("illegal object entity, must register entity first")
		return
	}

	ret = objModel
	return
}

// GetTypeModel GetTypeModel
func (s *providerImpl) GetTypeModel(vType model.Type) (ret model.Model, err error) {
	vType = vType.Depend()
	if vType == nil {
		return
	}
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

	ret = typeModel
	return
}

// GetValueModel GetValueModel
func (s *providerImpl) GetValueModel(modelVal reflect.Value) (ret model.Model, err error) {
	vType, vErr := s.getTypeFunc(modelVal.Type())
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

	ret = typeModel
	return
}

// GetSliceValueModel GetSliceValueModel
func (s *providerImpl) GetSliceValueModel(sliceVal reflect.Value) (ret model.Model, err error) {
	vType, vErr := s.getTypeFunc(sliceVal.Type())
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
	}

	if util.IsStructType(vType.GetValue()) {
		ret, err = s.getStructValue(vType, vVal)
		return
	}

	vType = vType.Depend()
	if util.IsBasicType(vType.GetValue()) {
		ret, err = helper.EncodeSliceValue(vVal.Get())
		return
	}

	ret, err = s.getSliceStructValue(vType, vVal)
	return
}

func (s *providerImpl) getStructValue(vType model.Type, vVal model.Value) (ret string, err error) {
	typeModel, typeErr := s.GetTypeModel(vType)
	if typeErr != nil {
		err = typeErr
		return
	}

	pkField := typeModel.GetPrimaryField()
	return getBasicValue(pkField.GetType(), vVal.Get())
}

func (s *providerImpl) getSliceStructValue(vType model.Type, vVal model.Value) (ret string, err error) {
	typeModel, typeErr := s.GetTypeModel(vType)
	if typeErr != nil {
		err = typeErr
		return
	}

	pkField := typeModel.GetPrimaryField()

	val := reflect.Indirect(vVal.Get())
	if val.Kind() != reflect.Slice {
		err = fmt.Errorf("illegal slice value")
		return
	}

	var sliceVal []string
	for idx := 0; idx < val.Len(); idx++ {
		strVal, strErr := getBasicValue(pkField.GetType(), val.Index(idx))
		if strErr != nil {
			err = strErr
			log.Printf("getStructValue failed, err:%s", err.Error())
			return
		}

		sliceVal = append(sliceVal, strVal)
	}

	ret = strings.Join(sliceVal, ",")
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

// GetValueDepend GetValue depend values
func (s *providerImpl) GetDependValue(vValue model.Value) (ret []reflect.Value, err error) {
	if vValue.IsNil() {
		return
	}

	val := vValue.Get()
	vType, vErr := s.getTypeFunc(val.Type())
	if vErr != nil {
		err = vErr
		return
	}
	if vType.Depend() == nil {
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

	if vType.GetValue() == util.TypeSliceField {
		val = reflect.Indirect(val)
		for idx := 0; idx < val.Len(); idx++ {
			ret = append(ret, val.Index(idx))
		}

		return
	}

	ret = append(ret, val)
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
	return &providerImpl{
		owner:             owner,
		localProvider:     true,
		modelCache:        model.NewCache(),
		getTypeFunc:       local.GetType,
		getValueModelFunc: local.GetValueModel,
		getTypeModelFunc:  local.GetTypeModel,
	}
}

// NewRemoteProvider model provider
func NewRemoteProvider(owner string) Provider {
	return &providerImpl{
		owner:             owner,
		localProvider:     false,
		modelCache:        model.NewCache(),
		getTypeFunc:       remote.GetType,
		getValueModelFunc: remote.GetValueModel,
		getTypeModelFunc:  remote.GetTypeModel,
	}
}
