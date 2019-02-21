package local

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// getSliceValueStr get slice value str
func getSliceValueStr(val reflect.Value, cache Cache) (ret string, err error) {
	valSlice := []interface{}{}

	rawVal := reflect.Indirect(val)
	pos := rawVal.Len()
	for idx := 0; idx < pos; {
		sv := rawVal.Index(idx)
		sv = reflect.Indirect(sv)
		switch sv.Kind() {
		case reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64,
			reflect.String:
			valSlice = append(valSlice, sv.Interface())
		case reflect.Struct:
			if sv.Type().String() == "time.Time" {
				datetimeStr, datetimeErr := getDateTimeValueStr(sv)
				if datetimeErr != nil {
					err = datetimeErr
					return
				}

				valSlice = append(valSlice, datetimeStr)
			} else {
				err = fmt.Errorf("no support slice element type, [%s]", sv.Type().String())
			}
		default:
			err = fmt.Errorf("no support slice element type, [%s]", sv.Type().String())
		}

		idx++
	}

	data, dataErr := json.Marshal(valSlice)
	if dataErr != nil {
		err = dataErr
	}
	ret = fmt.Sprintf("%s", string(data))

	return
}
