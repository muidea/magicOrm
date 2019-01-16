package orm

import (
	"fmt"
	"reflect"
	"strings"

	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/util"
)

type sliceValue struct {
	value          reflect.Value
	modelInfoCache model.StructInfoCache
}

func (s *sliceValue) String() (ret string, err error) {
	if s.value.Kind() != reflect.Slice {
		err = fmt.Errorf("illegal value type:%s", s.value.Type().String())
		return
	}

	retArray := []string{}
	len := s.value.Len()
	for idx := 0; idx < len; idx++ {
		val := s.value.Index(idx)
		fval, fErr := util.GetTypeValueEnum(reflect.Indirect(val).Type())
		if fErr != nil {
			err = fErr
			return
		}

		if util.IsBasicType(fval) {
			basicVal := &basicValue{value: val}

			strVal, strErr := basicVal.String()
			if strErr != nil {
				err = strErr
				return
			}

			retArray = append(retArray, strVal)
			continue
		}

		if util.IsStructType(fval) {
			structVal := &structValue{value: val, modelInfoCache: s.modelInfoCache}

			strVal, strErr := structVal.String()
			if strErr != nil {
				err = strErr
				return
			}

			retArray = append(retArray, strVal)
			continue
		}

		err = fmt.Errorf("illegal slice element type, type:%s", val.Type().String())
		return
	}

	ret = strings.Join(retArray, ",")
	return
}
