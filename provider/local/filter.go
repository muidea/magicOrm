package local

import (
	"fmt"
	"reflect"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/foundation/util"

	om "github.com/muidea/magicOrm/model"
	pu "github.com/muidea/magicOrm/provider/util"
)

type filterItem struct {
	oprCode om.OprCode
	value   *ValueImpl
}

func (s *filterItem) OprCode() om.OprCode {
	return s.oprCode
}

func (s *filterItem) OprValue() om.Value {
	return s.value
}

type filter struct {
	bindValue  *ValueImpl
	params     map[string]*filterItem
	maskValue  *ValueImpl
	pageFilter *util.Pagination
	sortFilter *util.SortFilter
}

func newFilter(valuePtr *ValueImpl) *filter {
	return &filter{bindValue: valuePtr, params: map[string]*filterItem{}}
}

func (s *filter) Equal(key string, val any) (err *cd.Result) {
	if val == nil {
		err = cd.NewResult(cd.IllegalParam, "illegal equal value")
		return
	}

	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := pu.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		log.Errorf("Equal failed, illegal value type, err:%s", err.Error())
		return
	}
	if om.IsSliceType(qvType) {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("equal failed, illegal value type, type:%s", qv.Type().String()))
		log.Errorf("Equal failed, err:%v", err.Error())
		return
	}

	//s.equalFilter = append(s.equalFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: om.EqualOpr, value: NewValue(qv)}
	return
}

func (s *filter) NotEqual(key string, val any) (err *cd.Result) {
	if val == nil {
		err = cd.NewResult(cd.IllegalParam, "illegal not equal value")
		return
	}

	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := pu.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		log.Errorf("NotEqual failed, illegal value type, err:%s", err.Error())
		return
	}
	if om.IsSliceType(qvType) {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("NotEqual failed, illegal value type, type:%s", qv.Type().String()))
		log.Errorf("NotEqual failed, err:%v", err.Error())
		return
	}

	//s.notEqualFilter = append(s.notEqualFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: om.NotEqualOpr, value: NewValue(qv)}
	return
}

func (s *filter) Below(key string, val any) (err *cd.Result) {
	if val == nil {
		err = cd.NewResult(cd.IllegalParam, "illegal below value")
		return
	}

	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := pu.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		log.Errorf("Below failed, illegal value type, err:%s", err.Error())
		return
	}
	if !om.IsBasicType(qvType) {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("below failed, illegal value type, type:%s", qv.Type().String()))
		log.Errorf("Below failed, err:%v", err.Error())
		return
	}

	//s.belowFilter = append(s.belowFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: om.BelowOpr, value: NewValue(qv)}
	return
}

func (s *filter) Above(key string, val any) (err *cd.Result) {
	if val == nil {
		err = cd.NewResult(cd.IllegalParam, "illegal above value")
		return
	}

	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := pu.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		log.Errorf("Above failed, illegal value type, err:%s", err.Error())
		return
	}
	if !om.IsBasicType(qvType) {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("above failed, illegal value type, type:%s", qv.Type().String()))
		log.Errorf("Above failed, err:%v", err.Error())
		return
	}

	//s.aboveFilter = append(s.aboveFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: om.AboveOpr, value: NewValue(qv)}
	return
}

func (s *filter) In(key string, val any) (err *cd.Result) {
	if val == nil {
		err = cd.NewResult(cd.IllegalParam, "illegal in value")
		return
	}

	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := pu.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		log.Errorf("In failed, illegal value type, err:%s", err.Error())
		return
	}
	if !om.IsSliceType(qvType) {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("in failed, illegal value type, type:%s", qv.Type().String()))
		log.Errorf("In failed, err:%v", err.Error())
		return
	}

	//s.inFilter = append(s.inFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: om.InOpr, value: NewValue(qv)}
	return
}

func (s *filter) NotIn(key string, val any) (err *cd.Result) {
	if val == nil {
		err = cd.NewResult(cd.IllegalParam, "illegal not in value")
		return
	}

	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := pu.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		log.Errorf("NotIn failed, illegal value type, err:%s", err.Error())
		return
	}
	if !om.IsSliceType(qvType) {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("notIn failed, illegal value type, type:%s", qv.Type().String()))
		log.Errorf("NotIn failed, err:%v", err.Error())
		return
	}

	//s.notInFilter = append(s.notInFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: om.NotInOpr, value: NewValue(qv)}
	return
}

func (s *filter) Like(key string, val any) (err *cd.Result) {
	if val == nil {
		err = cd.NewResult(cd.IllegalParam, "illegal like value")
		return
	}

	qv := reflect.Indirect(reflect.ValueOf(val))
	if qv.Kind() != reflect.String {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("like failed, illegal value type, type:%s", qv.Type().String()))
		log.Errorf("Like failed, illegal value type, err:%s", err.Error())
		return
	}

	//s.likeFilter = append(s.likeFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: om.LikeOpr, value: NewValue(qv)}
	return
}

func (s *filter) Page(pageFilter *util.Pagination) {
	s.pageFilter = pageFilter
}

func (s *filter) Sort(sorter *util.SortFilter) {
	s.sortFilter = sorter
}

func (s *filter) ValueMask(val any) (err *cd.Result) {
	if val == nil {
		err = cd.NewResult(cd.IllegalParam, "illegal value mask")
		return
	}

	qv := reflect.Indirect(reflect.ValueOf(val))
	bindType := reflect.Indirect(s.bindValue.value).Type().String()
	maskType := reflect.Indirect(qv).Type().String()
	if bindType != maskType {
		err = cd.NewResult(cd.UnExpected, fmt.Sprintf("mismatch mask value, bindType:%v, maskType:%v", bindType, maskType))
		log.Errorf("ValueMask failed, err:%v", err.Error())
		return
	}

	s.maskValue = NewValue(qv)
	return
}

func (s *filter) GetFilterItem(key string) om.FilterItem {
	v, ok := s.params[key]
	if ok {
		return v
	}

	return nil
}

func (s *filter) Pagination() (limit, offset int, paging bool) {
	paging = false
	if s.pageFilter == nil {
		return
	}

	paging = true
	limit = s.pageFilter.PageSize
	offset = s.pageFilter.PageSize * (s.pageFilter.PageNum - 1)
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 100
	}

	return
}

func (s *filter) Sorter() om.Sorter {
	if s.sortFilter == nil {
		return nil
	}

	return s.sortFilter
}

func (s *filter) MaskModel() om.Model {
	maskVal := s.bindValue
	if s.maskValue != nil {
		maskVal = s.maskValue
	}

	objPtr, objErr := getValueModel(maskVal.value)
	if objErr != nil {
		log.Errorf("MaskModel failed, getValueModel error:%s", objErr.Error())
		return nil
	}

	return objPtr.Copy(false)
}
