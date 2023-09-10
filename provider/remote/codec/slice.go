package codec

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
)

// encodeSlice get slice value str
func (s *impl) encodeSlice(vVal model.Value, vType model.Type) (ret string, err error) {
	vals, valErr := s.elemDependValue(vVal)
	if valErr != nil {
		err = valErr
		return
	}
	if len(vals) == 0 {
		return
	}
	if len(vals) == 1 {
		strVal, strErr := s.Encode(vals[0], vType.Elem())
		if strErr != nil {
			err = strErr
			return
		}

		ret = fmt.Sprintf("%v", strVal)
		return
	}

	items := []interface{}{}
	for _, val := range vals {
		strVal, strErr := s.Encode(val, vType.Elem())
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

func (s *impl) decodeStringSlice(val string, vType model.Type) (ret model.Value, err error) {
	tVal := vType.Interface()
	if val != "" {
		sliceVal := reflect.ValueOf(tVal.Get())
		if val[0] != '[' {
			itemVal, itemErr := s.Decode(val, vType.Elem())
			if itemErr != nil {
				err = itemErr
				return
			}

			sliceVal = reflect.Append(sliceVal, reflect.ValueOf(itemVal.Get()))
		} else {
			items := []any{}
			err = json.Unmarshal([]byte(val), &items)
			if err != nil {
				return
			}

			for idx := range items {
				itemVal, itemErr := s.Decode(items[idx], vType.Elem())
				if itemErr != nil {
					err = itemErr
					return
				}

				sliceVal = reflect.Append(sliceVal, reflect.ValueOf(itemVal.Get()))
			}
		}

		tVal.Set(sliceVal.Interface())
	}

	ret = tVal
	return
}

// decodeSlice decode slice from string
func (s *impl) decodeSlice(val interface{}, vType model.Type) (ret model.Value, err error) {
	var strVal string
	switch val.(type) {
	case string:
		strVal = val.(string)
	default:
		err = fmt.Errorf("decodeSlice failed, illegal slice value, val:%v", val)
	}
	if err != nil {
		return
	}

	ret, err = s.decodeStringSlice(strVal, vType)
	return
}
