package local

import (
	"fmt"
	"log"
	"reflect"

	"muidea.com/magicOrm/model"
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
			fv := field.GetValue()
			err = fv.Set(val)
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
			fv := field.GetValue()
			err = fv.Set(val)
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

// getTypeModel getTypeModel
func getTypeModel(modelType reflect.Type, cache Cache) (ret *modelImpl, err error) {
	rawType := modelType
	isPtr := false
	if rawType.Kind() == reflect.Ptr {
		rawType = rawType.Elem()
		isPtr = true
	}

	if rawType.Kind() != reflect.Struct {
		err = fmt.Errorf("illegal modelType, type:%s", rawType.String())
		return
	}

	if rawType.String() == "time.Time" {
		err = fmt.Errorf("illegal modelType, type:%s", rawType.String())
		return
	}
	modelInfo := cache.Fetch(rawType.Name())
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

	modelImpl := &modelImpl{modelType: rawType, fields: make([]*fieldImpl, 0)}
	fieldNum := rawType.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldType := rawType.Field(idx)
		fieldInfo, fieldErr := getFieldInfo(idx, fieldType)
		if fieldErr != nil {
			err = fieldErr
			log.Printf("getFieldInfo failed, name:%s, err:%s", fieldType.Name, err.Error())
			return
		}

		if fieldInfo != nil {
			modelImpl.fields = append(modelImpl.fields, fieldInfo)
		}
	}

	if len(modelImpl.fields) == 0 {
		err = fmt.Errorf("no define orm field, struct name:%s", modelImpl.GetName())
		return
	}

	cache.Put(modelImpl.GetName(), modelImpl)

	ret = modelImpl
	ret.isTypePtr = isPtr

	return
}

// getValueModel getValueModel
func getValueModel(modelVal reflect.Value, cache Cache) (ret *modelImpl, err error) {
	rawVal := modelVal
	if rawVal.Kind() == reflect.Ptr {
		if rawVal.IsNil() {
			err = fmt.Errorf("can't get value model from nil ptr")
			return
		}

		rawVal = reflect.Indirect(rawVal)
	}

	modelInfo, modelErr := getTypeModel(modelVal.Type(), cache)
	if modelErr != nil {
		err = modelErr
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

// getValueStr get value str
func getValueStr(vType model.Type, vVal model.Value, cache Cache) (ret string, err error) {
	rawType := vType.GetType()
	if rawType.Kind() == reflect.Ptr {
		rawType = rawType.Elem()
	}

	switch rawType.Kind() {
	case reflect.Bool:
		ret, err = getBoolValueStr(vVal.Get())
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		ret, err = getIntValueStr(vVal.Get())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		ret, err = getUintValueStr(vVal.Get())
	case reflect.Float32, reflect.Float64:
		ret, err = getFloatValueStr(vVal.Get())
	case reflect.String:
		strRet, strErr := getStringValueStr(vVal.Get())
		if strErr != nil {
			err = strErr
			return
		}
		ret = fmt.Sprintf("'%s'", strRet)
	case reflect.Slice:
		strRet, strErr := getSliceValueStr(vVal.Get(), cache)
		if strErr != nil {
			err = strErr
			return
		}
		ret = fmt.Sprintf("'%s'", strRet)
	case reflect.Struct:
		if rawType.String() == "time.Time" {
			strRet, strErr := getDateTimeValueStr(vVal.Get())
			if strErr != nil {
				err = strErr
				return
			}
			ret = fmt.Sprintf("'%s'", strRet)
		} else {
			ret, err = getStructValueStr(vVal.Get(), cache)
		}
	default:
		err = fmt.Errorf("illegal value kind, kind:%v", rawType.Kind())
	}
	return
}
