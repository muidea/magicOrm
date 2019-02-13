package remote

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"muidea.com/magicOrm/model"
)

// Provider remote provider
type Provider struct {
	modelCache model.Cache
}

// New create remote provider
func New(cache model.Cache) *Provider {
	return &Provider{modelCache: cache}
}

// GetObjectModel GetObjectModel
func (s *Provider) GetObjectModel(objPtr interface{}) (ret model.Model, err error) {
	info, err := GetInfo(objPtr)
	if err != nil {
		return
	}

	ret = info
	return
}

// GetTypeModel GetTypeModel
func (s *Provider) GetTypeModel(modelType reflect.Type) (ret model.Model, err error) {
	return
}

// GetValueModel GetValueModel
func (s *Provider) GetValueModel(modelVal reflect.Value) (ret model.Model, err error) {
	return
}

// GetValueStr GetValueStr
func (s *Provider) GetValueStr(value reflect.Value) (ret string, err error) {
	fValue := ""
	switch value.Kind() {
	case reflect.Slice:
		fValue, err = getSliceValStr(value)
	case reflect.Struct:
		fValue, err = getStructValStr(value)
	default:
		fValue, err = getBasicValStr(value)
	}
	if err != nil {
		return
	}

	ret = fValue
	return
}

// Reset Reset
func (s *Provider) Reset() {
	s.modelCache.Reset()
}

func getBasicValStr(value reflect.Value) (ret string, err error) {
	switch value.Kind() {
	case reflect.Slice, reflect.Struct:
		err = fmt.Errorf("illegal basic type, type:%s", value.Type().String())
	case reflect.Bool:
		if value.Bool() {
			ret = "1"
		} else {
			ret = "0"
		}
	case reflect.String:
		ret = fmt.Sprintf("'%v'", value.Interface())
	default:
		ret = fmt.Sprintf("%v", value.Interface())
	}

	return
}

func getStructValStr(value reflect.Value) (ret string, err error) {
	switch value.Kind() {
	case reflect.Struct:
		if value.Type().String() == "time.Time" {
			ret = value.Interface().(time.Time).Format("2006-01-02 15:04:05")
			ret = fmt.Sprintf("'%s'", ret)
		} else {
			//ret, err = GetModelValueStr(value)
		}
	default:
		err = fmt.Errorf("illegal struct type, type:%s", value.Type().String())
	}

	return
}

func getSliceValStr(value reflect.Value) (ret string, err error) {
	valSlice := []string{}
	pos := value.Len()
	for idx := 0; idx < pos; {
		sv := value.Index(idx)
		sv = reflect.Indirect(sv)
		strVal := ""
		switch sv.Kind() {
		case reflect.Slice:
			err = fmt.Errorf("illegal slice type, type:%s", value.Type().String())
		case reflect.Struct:
			strVal, err = getStructValStr(sv)
		default:
			strVal, err = getBasicValStr(sv)
		}

		if err != nil {
			return
		}

		valSlice = append(valSlice, strVal)
		idx++
	}

	ret = strings.Join(valSlice, ",")
	return
}
