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
	if len(vals) == 0 {
		return
	}
	if len(vals) == 1 {
		strVal, strErr := s.Encode(vals[0], tType.Elem())
		if strErr != nil {
			err = strErr
			return
		}

		ret = fmt.Sprintf("%v", strVal)
		return
	}

	items := []interface{}{}
	for _, val := range vals {
		strVal, strErr := s.Encode(val, tType.Elem())
		if strErr != nil {
			err = strErr
			return
		}

		items = append(items, strVal)
	}

	if len(items) > 0 {
		data, dataErr := json.Marshal(items)
		if dataErr != nil {
			err = dataErr
			return
		}

		ret = string(data)
	}
	return
}

func (s *impl) decodeStringSlice(val string, tType model.Type) (ret model.Value, err error) {
	tVal := tType.Interface()
	if val != "" {
		sliceVal := tVal.Get()
		if val[0] != '[' {
			itemVal, itemErr := s.Decode(val, tType.Elem())
			if itemErr != nil {
				err = itemErr
				return
			}

			sliceVal = reflect.Append(sliceVal, itemVal.Get())
		} else {
			items := []interface{}{}
			err = json.Unmarshal([]byte(val), &items)
			if err != nil {
				return
			}

			for idx := range items {
				itemVal, itemErr := s.Decode(items[idx], tType.Elem())
				if itemErr != nil {
					err = itemErr
					return
				}

				sliceVal = reflect.Append(sliceVal, itemVal.Get())
			}
		}
		tVal.Set(sliceVal)
	}

	ret = tVal
	return
}

func (s *impl) decodeReflectSlice(val reflect.Value, tType model.Type) (ret model.Value, err error) {
	tVal := tType.Interface()
	sliceVal := tVal.Get()
	for idx := 0; idx < val.Len(); idx++ {
		v := val.Index(idx)
		itemVal, itemErr := s.Decode(v.Interface(), tType.Elem())
		if itemErr != nil {
			err = itemErr
			return
		}

		sliceVal = reflect.Append(sliceVal, itemVal.Get())
	}
	tVal.Set(sliceVal)
	ret = tVal
	return
}

// decodeSlice decode slice from string
func (s *impl) decodeSlice(val interface{}, tType model.Type) (ret model.Value, err error) {
	rVal := reflect.ValueOf(val)
	if rVal.Kind() == reflect.Interface {
		rVal = rVal.Elem()
	}
	rVal = reflect.Indirect(rVal)

	switch rVal.Kind() {
	case reflect.String:
		ret, err = s.decodeStringSlice(rVal.String(), tType)
	case reflect.Slice:
		ret, err = s.decodeReflectSlice(rVal, tType)
	default:
		err = fmt.Errorf("illegal slice value, val:%v", val)
	}
	return
}
