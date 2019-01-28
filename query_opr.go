package orm

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"muidea.com/magicOrm/local"
	"muidea.com/magicOrm/model"
)

func getSliceValStr(val model.FieldValue) (ret string, err error) {
	value, valueErr := val.Get()
	if valueErr != nil {
		err = valueErr
		return
	}

	if value.Kind() != reflect.Slice {
		err = fmt.Errorf("illegal value, type:%s", value.Type().String())
		return
	}

	valSlice := []string{}
	pos := value.Len()
	for idx := 0; idx < pos; {
		sv := value.Index(idx)
		if sv.Kind() != reflect.Ptr {
			sv = sv.Addr()
		}

		sfieldVal, sfieldErr := local.NewFieldValue(sv)
		if sfieldErr != nil {
			err = sfieldErr
			return
		}

		strVal, strErr := sfieldVal.ValueStr()
		if strErr != nil {
			err = strErr
		}

		valSlice = append(valSlice, strVal)
		idx++
	}

	ret = strings.Join(valSlice, ",")
	return
}

func equleOpr(name string, value model.FieldValue) (ret string, err error) {
	val, valErr := value.ValueStr()
	if valErr == nil {
		ret = fmt.Sprintf("`%s` = %s", name, val)
		return
	}
	err = valErr

	log.Printf("get value string failed, err:%s", err.Error())
	return
}

func notEquleOpr(name string, value model.FieldValue) (ret string, err error) {
	val, valErr := value.ValueStr()
	if valErr == nil {
		ret = fmt.Sprintf("`%s` != %s", name, val)
		return
	}
	err = valErr

	log.Printf("get value string failed, err:%s", err.Error())
	return
}

func belowOpr(name string, value model.FieldValue) (ret string, err error) {
	val, valErr := value.ValueStr()
	if valErr == nil {
		ret = fmt.Sprintf("`%s` < %s", name, val)
		return
	}
	err = valErr

	log.Printf("get value string failed, err:%s", err.Error())
	return
}

func aboveOpr(name string, value model.FieldValue) (ret string, err error) {
	val, valErr := value.ValueStr()
	if valErr == nil {
		ret = fmt.Sprintf("`%s` > %s", name, val)
		return
	}
	err = valErr

	log.Printf("get value string failed, err:%s", err.Error())
	return
}

func inOpr(name string, value model.FieldValue) (ret string, err error) {
	val, valErr := getSliceValStr(value)
	if valErr == nil {
		if val != "" {
			ret = fmt.Sprintf("`%s` in (%v)", name, val)
		}
		return
	}
	err = valErr

	log.Printf("get value string failed, err:%s", err.Error())
	return
}

func notInOpr(name string, value model.FieldValue) (ret string, err error) {
	val, valErr := getSliceValStr(value)
	if valErr == nil {
		if val != "" {
			ret = fmt.Sprintf("`%s` not in (%v)", name, val)
		}

		return
	}
	err = valErr

	log.Printf("get value string failed, err:%s", err.Error())
	return
}

func likeOpr(name string, value model.FieldValue) (ret string, err error) {
	val, valErr := value.ValueStr()
	if valErr == nil {
		val := val[1 : len(val)-1]
		if val != "" {
			ret = fmt.Sprintf("`%s` LIKE '%%%s%%'", name, val)
		}

		return
	}
	err = valErr

	log.Printf("get value string failed, err:%s", err.Error())
	return
}
