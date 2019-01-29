package mysql

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

// EquleOpr EquleOpr
func EquleOpr(name string, value model.FieldValue) (ret string, err error) {
	val, valErr := value.ValueStr()
	if valErr == nil {
		ret = fmt.Sprintf("`%s` = %s", name, val)
		return
	}
	err = valErr

	log.Printf("get value string failed, err:%s", err.Error())
	return
}

// NotEquleOpr NotEquleOpr
func NotEquleOpr(name string, value model.FieldValue) (ret string, err error) {
	val, valErr := value.ValueStr()
	if valErr == nil {
		ret = fmt.Sprintf("`%s` != %s", name, val)
		return
	}
	err = valErr

	log.Printf("get value string failed, err:%s", err.Error())
	return
}

// BelowOpr BelowOpr
func BelowOpr(name string, value model.FieldValue) (ret string, err error) {
	val, valErr := value.ValueStr()
	if valErr == nil {
		ret = fmt.Sprintf("`%s` < %s", name, val)
		return
	}
	err = valErr

	log.Printf("get value string failed, err:%s", err.Error())
	return
}

// AboveOpr AboveOpr
func AboveOpr(name string, value model.FieldValue) (ret string, err error) {
	val, valErr := value.ValueStr()
	if valErr == nil {
		ret = fmt.Sprintf("`%s` > %s", name, val)
		return
	}
	err = valErr

	log.Printf("get value string failed, err:%s", err.Error())
	return
}

// InOpr InOpr
func InOpr(name string, value model.FieldValue) (ret string, err error) {
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

// NotInOpr NotInOpr
func NotInOpr(name string, value model.FieldValue) (ret string, err error) {
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

// LikeOpr LikeOpr
func LikeOpr(name string, value model.FieldValue) (ret string, err error) {
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
