package orm

import (
	"fmt"
	"log"
	"reflect"

	"muidea.com/magicCommon/foundation/util"
	"muidea.com/magicOrm/model"
	ormutil "muidea.com/magicOrm/util"
)

// queryFilter queryFilter
type queryFilter struct {
	params         map[string]opr
	modelInfoCache model.StructInfoCache
}

type queryValue interface {
	String() (string, error)
}

func newQueryValue(qv interface{}, cache model.StructInfoCache) (ret queryValue, err error) {
	val := reflect.Indirect(reflect.ValueOf(qv))
	fval, fErr := ormutil.GetTypeValueEnum(val.Type())
	if fErr != nil {
		err = fErr
		return
	}

	if ormutil.IsBasicType(fval) {
		ret = &basicValue{value: val}
		return
	}

	if ormutil.IsStructType(fval) {
		ret = &structValue{value: val, modelInfoCache: cache}
		return
	}

	if ormutil.IsSliceType(fval) {
		ret = &sliceValue{value: val, modelInfoCache: cache}
		return
	}

	err = fmt.Errorf("illegal query value, type:%s", val.Type().String())
	return
}

func (s *queryFilter) Equle(key string, val interface{}) (err error) {
	value, valErr := newQueryValue(val, s.modelInfoCache)
	if valErr != nil {
		err = valErr
		return
	}

	s.params[key] = &equleOpr{name: key, value: value}
	return
}

func (s *queryFilter) NotEqule(key string, val interface{}) (err error) {
	value, valErr := newQueryValue(val, s.modelInfoCache)
	if valErr != nil {
		err = valErr
		return
	}

	s.params[key] = &notEquleOpr{name: key, value: value}
	return
}

func (s *queryFilter) Below(key string, val interface{}) (err error) {
	value, valErr := newQueryValue(val, s.modelInfoCache)
	if valErr != nil {
		err = valErr
		return
	}

	s.params[key] = &belowOpr{name: key, value: value}
	return
}

func (s *queryFilter) Above(key string, val interface{}) (err error) {
	value, valErr := newQueryValue(val, s.modelInfoCache)
	if valErr != nil {
		err = valErr
		return
	}

	s.params[key] = &aboveOpr{name: key, value: value}
	return
}

func (s *queryFilter) In(key string, val []interface{}) (err error) {
	value, valErr := newQueryValue(val, s.modelInfoCache)
	if valErr != nil {
		err = valErr
		return
	}

	s.params[key] = &inOpr{name: key, value: value}
	return
}

func (s *queryFilter) NotIn(key string, val []interface{}) (err error) {
	value, valErr := newQueryValue(val, s.modelInfoCache)
	if valErr != nil {
		err = valErr
		return
	}

	s.params[key] = &notInOpr{name: key, value: value}
	return
}

func (s *queryFilter) PageFilter(filter *util.PageFilter) {
}

func (s *queryFilter) Builder(structInfo model.StructInfo) (ret string, err error) {
	if structInfo == nil {
		return
	}

	fields := structInfo.GetFields()
	for _, field := range *fields {
		fType := field.GetFieldType()
		fDepend, _ := fType.Depend()
		if fDepend != nil {
			continue
		}

		val, ok := s.params[field.GetFieldName()]
		if !ok {
			continue
		}
		strVal, strErr := val.String()
		if strErr != nil {
			err = strErr
			return
		}

		if ret == "" {
			ret = fmt.Sprintf("%s", strVal)
		} else {
			ret = fmt.Sprintf("%s AND %s", ret, strVal)
		}
	}

	return
}

type opr interface {
	String() (string, error)
}

type equleOpr struct {
	name  string
	value queryValue
}

func (s *equleOpr) String() (ret string, err error) {
	val, valErr := s.value.String()
	if valErr == nil {
		ret = fmt.Sprintf("`%s` = %s", s.name, val)
		return
	}
	err = valErr

	log.Printf("get value string failed, err:%s", err.Error())
	return
}

type notEquleOpr struct {
	name  string
	value queryValue
}

func (s *notEquleOpr) String() (ret string, err error) {
	val, valErr := s.value.String()
	if valErr == nil {
		ret = fmt.Sprintf("`%s` != %s", s.name, val)
		return
	}
	err = valErr

	log.Printf("get value string failed, err:%s", err.Error())
	return
}

type belowOpr struct {
	name  string
	value queryValue
}

func (s *belowOpr) String() (ret string, err error) {
	val, valErr := s.value.String()
	if valErr == nil {
		ret = fmt.Sprintf("`%s` < %s", s.name, val)
		return
	}
	err = valErr

	log.Printf("get value string failed, err:%s", err.Error())
	return
}

type aboveOpr struct {
	name  string
	value queryValue
}

func (s *aboveOpr) String() (ret string, err error) {
	val, valErr := s.value.String()
	if valErr == nil {
		ret = fmt.Sprintf("`%s` > %s", s.name, val)
		return
	}
	err = valErr

	log.Printf("get value string failed, err:%s", err.Error())
	return
}

type inOpr struct {
	name  string
	value queryValue
}

func (s *inOpr) String() (ret string, err error) {
	val, valErr := s.value.String()
	if valErr == nil {
		ret = fmt.Sprintf("`%s` in (%v)", s.name, val)
		return
	}
	err = valErr

	log.Printf("get value string failed, err:%s", err.Error())
	return
}

type notInOpr struct {
	name  string
	value queryValue
}

func (s *notInOpr) String() (ret string, err error) {
	val, valErr := s.value.String()
	if valErr == nil {
		ret = fmt.Sprintf("`%s` not in (%v)", s.name, val)
		return
	}
	err = valErr

	log.Printf("get value string failed, err:%s", err.Error())
	return
}
