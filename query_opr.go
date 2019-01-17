package orm

import (
	"fmt"
	"log"

	"muidea.com/magicOrm/model"
)

func equleOpr(name string, value model.FieldValue) (ret string, err error) {
	val, valErr := value.GetValueStr()
	if valErr == nil {
		ret = fmt.Sprintf("`%s` = %s", name, val)
		return
	}
	err = valErr

	log.Printf("get value string failed, err:%s", err.Error())
	return
}

func notEquleOpr(name string, value model.FieldValue) (ret string, err error) {
	val, valErr := value.GetValueStr()
	if valErr == nil {
		ret = fmt.Sprintf("`%s` != %s", name, val)
		return
	}
	err = valErr

	log.Printf("get value string failed, err:%s", err.Error())
	return
}

func belowOpr(name string, value model.FieldValue) (ret string, err error) {
	val, valErr := value.GetValueStr()
	if valErr == nil {
		ret = fmt.Sprintf("`%s` < %s", name, val)
		return
	}
	err = valErr

	log.Printf("get value string failed, err:%s", err.Error())
	return
}

func aboveOpr(name string, value model.FieldValue) (ret string, err error) {
	val, valErr := value.GetValueStr()
	if valErr == nil {
		ret = fmt.Sprintf("`%s` > %s", name, val)
		return
	}
	err = valErr

	log.Printf("get value string failed, err:%s", err.Error())
	return
}

func inOpr(name string, value model.FieldValue) (ret string, err error) {
	val, valErr := value.GetValueStr()
	if valErr == nil {
		ret = fmt.Sprintf("`%s` in (%v)", name, val)
		return
	}
	err = valErr

	log.Printf("get value string failed, err:%s", err.Error())
	return
}

func notInOpr(name string, value model.FieldValue) (ret string, err error) {
	val, valErr := value.GetValueStr()
	if valErr == nil {
		ret = fmt.Sprintf("`%s` not in (%v)", name, val)
		return
	}
	err = valErr

	log.Printf("get value string failed, err:%s", err.Error())
	return
}
