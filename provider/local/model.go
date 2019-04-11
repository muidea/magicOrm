package local

import (
	"fmt"
	"log"
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/util"
)

// modelImpl single model
type modelImpl struct {
	modelType reflect.Type
	isTypePtr bool
	fields    []*fieldImpl
}

func (s *modelImpl) GetName() string {
	return s.modelType.Name()
}

// GetPkgPath GetPkgPath
func (s *modelImpl) GetPkgPath() string {
	return s.modelType.PkgPath()
}

// GetFields GetFields
func (s *modelImpl) GetFields() (ret model.Fields) {
	for _, field := range s.fields {
		ret = append(ret, field)
	}

	return
}

// SetFieldValue SetFieldValue
func (s *modelImpl) SetFieldValue(idx int, val reflect.Value) (err error) {
	for _, field := range s.fields {
		if field.GetIndex() == idx {
			err = field.SetValue(val)
			return
		}
	}

	err = fmt.Errorf("out of index, index:%d", idx)
	return
}

// UpdateFieldValue UpdateFieldValue
func (s *modelImpl) UpdateFieldValue(name string, val reflect.Value) (err error) {
	for _, field := range s.fields {
		if field.GetName() == name {
			err = field.UpdateValue(val)
			return
		}
	}

	err = fmt.Errorf("no found field, name:%s", name)
	return
}

// GetPrimaryField GetPrimaryField
func (s *modelImpl) GetPrimaryField() (ret model.Field) {
	for _, field := range s.fields {
		if field.IsPrimary() {
			ret = field
			return
		}
	}

	return
}

func (s *modelImpl) IsPtrModel() bool {
	return s.isTypePtr
}

func (s *modelImpl) Interface() reflect.Value {
	retVal := reflect.New(s.modelType)
	if !s.isTypePtr {
		retVal = retVal.Elem()
	}

	return retVal
}

func (s *modelImpl) Copy() *modelImpl {
	modelInfo := &modelImpl{modelType: s.modelType, isTypePtr: s.isTypePtr, fields: []*fieldImpl{}}
	for _, field := range s.fields {
		modelInfo.fields = append(modelInfo.fields, field.Copy())
	}

	return modelInfo
}

// Dump Dump
func (s *modelImpl) Dump(cache Cache) (ret string) {
	ret = fmt.Sprintf("\nmodelImpl:\n")
	ret = fmt.Sprintf("%s\tname:%s, pkgPath:%s\n", ret, s.GetName(), s.GetPkgPath())

	ret = fmt.Sprintf("%sfields:\n", ret)
	for _, field := range s.fields {
		ret = fmt.Sprintf("%s\t%s\n", ret, field.Dump(cache))
	}

	log.Print(ret)

	return
}

// getObjectModel GetObjectModel
func getObjectModel(modelObj interface{}, cache Cache) (ret *modelImpl, err error) {
	modelVal := reflect.ValueOf(modelObj)

	ret, err = getValueModel(modelVal, cache)
	if err != nil {
		log.Printf("getValueModel failed, err:%s", err.Error())
		return
	}

	return
}

// getValueModel getValueModel
func getValueModel(modelVal reflect.Value, cache Cache) (ret *modelImpl, err error) {
	var vType model.Type
	vType, vErr := newType(modelVal.Type())
	if vErr != nil {
		err = vErr
		log.Printf("newType failed, err:%s", err.Error())
		return
	}

	if util.IsSliceType(vType.GetValue()) {
		err = fmt.Errorf("illegal value, type:%s", modelVal.Type().String())
		return
	}

	modelInfo, modelErr := getTypeModel(vType, cache)
	if modelErr != nil {
		err = modelErr
		log.Printf("getTypeModel failed, err:%s", err.Error())
		return
	}

	rawVal := modelVal
	if rawVal.Kind() == reflect.Ptr {
		if rawVal.IsNil() {
			return
		}

		rawVal = reflect.Indirect(rawVal)
	}

	if rawVal.Kind() != reflect.Struct {
		ret = modelInfo
		return
	}

	fieldNum := rawVal.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldVal := rawVal.Field(idx)
		err = modelInfo.SetFieldValue(idx, fieldVal)
		if err != nil {
			log.Printf("SetFieldValue failed, err:%s", err.Error())
			return
		}
	}

	ret = modelInfo

	return
}

// getSliceValueModel getSliceValueModel
func getSliceValueModel(modelVal reflect.Value, cache Cache) (ret *modelImpl, err error) {
	var vType model.Type
	vType, vErr := newType(modelVal.Type())
	if vErr != nil {
		err = vErr
		log.Printf("newType failed, err:%s", err.Error())
		return
	}

	if !util.IsSliceType(vType.GetValue()) {
		err = fmt.Errorf("illegal slice value")
		return
	}

	vType = vType.Elem()
	if !util.IsStructType(vType.GetValue()) {
		err = fmt.Errorf("illegal slice item value")
		return
	}

	modelInfo, modelErr := getTypeModel(vType, cache)
	if modelErr != nil {
		err = modelErr
		log.Printf("getTypeModel failed, err:%s", err.Error())
		return
	}

	ret = modelInfo

	return
}

// getTypeModel getTypeModel
func getTypeModel(vType model.Type, cache Cache) (ret *modelImpl, err error) {
	rawType := vType.GetType()
	isPtr := vType.IsPtrType()

	modelInfo := cache.Fetch(vType.GetName())
	if modelInfo != nil {
		preType := modelInfo.modelType
		if preType.PkgPath() != rawType.PkgPath() {
			err = fmt.Errorf("duplicate model info,name:%s", rawType.Name())
			return
		}

		ret = modelInfo.Copy()
		ret.isTypePtr = isPtr
		return
	}

	err = fmt.Errorf("can't find type model, typeName:%s", vType.GetName())

	return
}

// getValueStr get value str
func getValueStr(vType model.Type, vVal model.Value, cache Cache) (ret string, err error) {
	if vVal.IsNil() {
		return
	}

	rawType := vType.GetType()
	switch rawType.Kind() {
	case reflect.Bool:
		ret, err = helper.EncodeBoolValue(vVal.Get())
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		ret, err = helper.EncodeIntValue(vVal.Get())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		ret, err = helper.EncodeUintValue(vVal.Get())
	case reflect.Float32, reflect.Float64:
		ret, err = helper.EncodeFloatValue(vVal.Get())
	case reflect.String:
		strRet, strErr := helper.EncodeStringValue(vVal.Get())
		if strErr != nil {
			err = strErr
			return
		}
		ret = fmt.Sprintf("'%s'", strRet)
	case reflect.Slice:
		strRet, strErr := helper.EncodeSliceValue(vVal.Get())
		if strErr != nil {
			err = strErr
			return
		}
		ret = fmt.Sprintf("'%s'", strRet)
	case reflect.Struct:
		if rawType.String() == "time.Time" {
			strRet, strErr := helper.EncodeDateTimeValue(vVal.Get())
			if strErr != nil {
				err = strErr
				return
			}
			ret = fmt.Sprintf("'%s'", strRet)
		} else {
			ret, err = getStructValue(vVal.Get(), cache)
		}
	default:
		err = fmt.Errorf("illegal value kind, kind:%v", rawType.Kind())
	}

	return
}

// getStructValue get struct value str
func getStructValue(val reflect.Value, cache Cache) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	modelImpl, modelErr := getValueModel(rawVal, cache)
	if modelErr != nil {
		err = modelErr
		return
	}

	pk := modelImpl.GetPrimaryField()
	if pk == nil {
		err = fmt.Errorf("no define primary field")
		return
	}

	ret, err = getValueStr(pk.GetType(), pk.GetValue(), cache)

	return
}
