package codec

import (
	"encoding/json"
	"fmt"
	"reflect"

	cd "github.com/muidea/magicCommon/def"
	pu "github.com/muidea/magicOrm/provider/util"

	"github.com/muidea/magicOrm/model"
)

// encodeSlice get slice value str
func (s *impl) encodeSlice(vVal model.Value, vType model.Type) (ret string, err *cd.Result) {
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
			err = cd.NewError(cd.UnExpected, dataErr.Error())
			return
		}

		ret = string(data)
	}

	return
}

func (s *impl) decodeStringSlice(val string, vType model.Type) (ret model.Value, err *cd.Result) {
	tVal, _ := vType.Interface(nil)
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
			byteErr := json.Unmarshal([]byte(val), &items)
			if byteErr != nil {
				err = cd.NewError(cd.UnExpected, byteErr.Error())
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
func (s *impl) decodeSlice(val interface{}, vType model.Type) (ret model.Value, err *cd.Result) {
	strVal, strErr := pu.GetString(val)
	if strErr != nil {
		err = strErr
		return
	}

	ret, err = s.decodeStringSlice(strVal, vType)
	return
}
