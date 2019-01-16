package orm

import (
	"fmt"
	"reflect"
	"time"
)

type basicValue struct {
	value reflect.Value
}

func (s *basicValue) String() (ret string, err error) {
	switch s.value.Kind() {
	case reflect.Bool:
		if s.value.Bool() {
			ret = "1"
		} else {
			ret = "0"
		}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		ret = fmt.Sprintf("%d", s.value.Int())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		ret = fmt.Sprintf("%d", s.value.Uint())
	case reflect.Float32, reflect.Float64:
		ret = fmt.Sprintf("%f", s.value.Float())
	case reflect.String:
		ret = fmt.Sprintf("'%s'", s.value.String())
	case reflect.Struct:
		if s.value.Type().String() == "time.Time" {
			ts, _ := s.value.Interface().(time.Time)
			ret = fmt.Sprintf("'%s'", ts.Format("2006-01-02 15:04:05"))
		}
	default:
		err = fmt.Errorf("illegal value type:%s", s.value.Type().String())
		ret = ""
	}

	return
}
