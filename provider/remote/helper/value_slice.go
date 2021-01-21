package helper

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
)

// encodeSlice get slice value str
func (s *impl) encodeSlice(vVal model.Value, tType model.Type) (ret string, err error) {
	vals, valErr := s.elemDependValue(vVal)
	if valErr != nil {
		err = valErr
		return
	}
	items := []string{}
	for _, val := range vals {
		strVal, strErr := s.Encode(val, tType.Elem())
		if strErr != nil {
			err = strErr
			return
		}

		items = append(items, strVal)
	}

	data, dataErr := json.Marshal(items)
	if dataErr != nil {
		err = dataErr
		return
	}

	ret = string(data)
	return
}

func (s *impl) decodeStringSlice(tVal reflect.Value, tType model.Type, cVal reflect.Value) (ret reflect.Value, err error) {
	sVal, sErr := GetTypeValue(tType)
	if sErr != nil {
		err = sErr
		return
	}

	items := []string{}
	err = json.Unmarshal([]byte(tVal.String()), &items)
	if err != nil {
		return
	}

	eType := tType.Elem()
	for idx := range items {
		eVal, eErr := GetTypeValue(eType)
		if eErr != nil {
			err = eErr
			return
		}

		itemVal := reflect.ValueOf(items[idx])
		eVal, eErr = s.decodeInternal(itemVal, eType, eVal)
		if eErr != nil {
			err = eErr
			return
		}

		sVal = reflect.Append(sVal, eVal)
	}
	cVal.Set(sVal)

	ret = cVal
	return
}

func (s *impl) decodeReflectSlice(tVal reflect.Value, tType model.Type, cVal reflect.Value) (ret reflect.Value, err error) {
	sVal, sErr := GetTypeValue(tType)
	if sErr != nil {
		err = sErr
		return
	}

	eType := tType.Elem()
	for idx := 0; idx < tVal.Len(); idx++ {
		eVal, eErr := GetTypeValue(eType)
		if eErr != nil {
			err = eErr
			return
		}

		eVal, eErr = s.decodeInternal(tVal.Index(idx), eType, eVal)
		if eErr != nil {
			err = eErr
			return
		}

		sVal = reflect.Append(sVal, eVal)
	}
	cVal.Set(sVal)

	ret = cVal
	return
}

// decodeSlice decode slice from string
func (s *impl) decodeSlice(tVal reflect.Value, tType model.Type, cVal reflect.Value) (ret reflect.Value, err error) {
	switch tVal.Kind() {
	case reflect.String:
		ret, err = s.decodeStringSlice(tVal, tType, cVal)
	case reflect.Slice:
		ret, err = s.decodeReflectSlice(tVal, tType, cVal)
	default:
		err = fmt.Errorf("illegal slice value, value type:%v", tVal.Type().String())
	}

	return
}
