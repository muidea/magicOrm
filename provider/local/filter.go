package local

import (
	"fmt"
	"reflect"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/utils"
)

type filterItem struct {
	oprCode models.OprCode
	value   *ValueImpl
}

func (s *filterItem) OprCode() models.OprCode {
	return s.oprCode
}

func (s *filterItem) OprValue() models.Value {
	return s.value
}

type filter struct {
	bindValue  *ValueImpl
	params     map[string]*filterItem
	maskValue  *ValueImpl
	pageFilter *utils.Pagination
	sortFilter *utils.SortFilter
}

func newFilter(valuePtr *ValueImpl) *filter {
	return &filter{bindValue: valuePtr, params: map[string]*filterItem{}}
}

func (s *filter) GetName() string {
	return ""
}

func (s *filter) GetPkgPath() string {
	return ""
}

func (s *filter) Equal(key string, val any) (err *cd.Error) {
	if val == nil {
		err = cd.NewError(cd.IllegalParam, "illegal equal value")
		return
	}

	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := utils.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		log.Errorf("Equal failed, illegal value type, err:%s", err.Error())
		return
	}
	if models.IsSliceType(qvType) {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("equal failed, illegal value type, type:%s", qv.Type().String()))
		log.Errorf("Equal failed, err:%v", err.Error())
		return
	}

	//s.equalFilter = append(s.equalFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: models.EqualOpr, value: NewValue(qv)}
	return
}

func (s *filter) NotEqual(key string, val any) (err *cd.Error) {
	if val == nil {
		err = cd.NewError(cd.IllegalParam, "illegal not equal value")
		return
	}

	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := utils.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		log.Errorf("NotEqual failed, illegal value type, err:%s", err.Error())
		return
	}
	if models.IsSliceType(qvType) {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("NotEqual failed, illegal value type, type:%s", qv.Type().String()))
		log.Errorf("NotEqual failed, err:%v", err.Error())
		return
	}

	//s.notEqualFilter = append(s.notEqualFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: models.NotEqualOpr, value: NewValue(qv)}
	return
}

func (s *filter) Below(key string, val any) (err *cd.Error) {
	if val == nil {
		err = cd.NewError(cd.IllegalParam, "illegal below value")
		return
	}

	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := utils.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		log.Errorf("Below failed, illegal value type, err:%s", err.Error())
		return
	}
	if !models.IsBasicType(qvType) {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("below failed, illegal value type, type:%s", qv.Type().String()))
		log.Errorf("Below failed, err:%v", err.Error())
		return
	}

	//s.belowFilter = append(s.belowFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: models.BelowOpr, value: NewValue(qv)}
	return
}

func (s *filter) Above(key string, val any) (err *cd.Error) {
	if val == nil {
		err = cd.NewError(cd.IllegalParam, "illegal above value")
		return
	}

	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := utils.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		log.Errorf("Above failed, illegal value type, err:%s", err.Error())
		return
	}
	if !models.IsBasicType(qvType) {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("above failed, illegal value type, type:%s", qv.Type().String()))
		log.Errorf("Above failed, err:%v", err.Error())
		return
	}

	//s.aboveFilter = append(s.aboveFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: models.AboveOpr, value: NewValue(qv)}
	return
}

func (s *filter) In(key string, val any) (err *cd.Error) {
	if val == nil {
		err = cd.NewError(cd.IllegalParam, "illegal in value")
		return
	}

	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := utils.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		log.Errorf("In failed, illegal value type, err:%s", err.Error())
		return
	}
	if !models.IsSliceType(qvType) {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("in failed, illegal value type, type:%s", qv.Type().String()))
		log.Errorf("In failed, err:%v", err.Error())
		return
	}

	//s.inFilter = append(s.inFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: models.InOpr, value: NewValue(qv)}
	return
}

func (s *filter) NotIn(key string, val any) (err *cd.Error) {
	if val == nil {
		err = cd.NewError(cd.IllegalParam, "illegal not in value")
		return
	}

	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := utils.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		log.Errorf("NotIn failed, illegal value type, err:%s", err.Error())
		return
	}
	if !models.IsSliceType(qvType) {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("notIn failed, illegal value type, type:%s", qv.Type().String()))
		log.Errorf("NotIn failed, err:%v", err.Error())
		return
	}

	//s.notInFilter = append(s.notInFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: models.NotInOpr, value: NewValue(qv)}
	return
}

func (s *filter) Like(key string, val any) (err *cd.Error) {
	if val == nil {
		err = cd.NewError(cd.IllegalParam, "illegal like value")
		return
	}

	qv := reflect.Indirect(reflect.ValueOf(val))
	if qv.Kind() != reflect.String {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("like failed, illegal value type, type:%s", qv.Type().String()))
		log.Errorf("Like failed, illegal value type, err:%s", err.Error())
		return
	}

	//s.likeFilter = append(s.likeFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: models.LikeOpr, value: NewValue(qv)}
	return
}

func (s *filter) Pagination(pageNum, pageSize int64) {
	s.pageFilter = &utils.Pagination{
		PageNum:  pageNum,
		PageSize: pageSize,
	}
}

func (s *filter) Sort(fieldName string, ascFlag bool) {
	s.sortFilter = &utils.SortFilter{
		FieldName: fieldName,
		AscFlag:   ascFlag,
	}
}

func (s *filter) ValueMask(val any) (err *cd.Error) {
	if val == nil {
		err = cd.NewError(cd.IllegalParam, "illegal value mask")
		return
	}

	qv := reflect.Indirect(reflect.ValueOf(val))
	bindType := reflect.Indirect(s.bindValue.value).Type().String()
	maskType := reflect.Indirect(qv).Type().String()
	if bindType != maskType {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("mismatch mask value, bindType:%v, maskType:%v", bindType, maskType))
		log.Errorf("ValueMask failed, err:%v", err.Error())
		return
	}

	s.maskValue = NewValue(qv)
	return
}

func (s *filter) GetFilterItem(key string) models.FilterItem {
	v, ok := s.params[key]
	if ok {
		return v
	}

	return nil
}

func (s *filter) Paginationer() models.Paginationer {
	if s.pageFilter == nil {
		return nil
	}

	return s.pageFilter
}

func (s *filter) Sorter() models.Sorter {
	if s.sortFilter == nil {
		return nil
	}

	return s.sortFilter
}

func (s *filter) MaskModel() models.Model {
	maskVal := s.bindValue
	if s.maskValue != nil {
		maskVal = s.maskValue
	}

	objPtr, objErr := getValueModel(maskVal.value, models.OriginView)
	if objErr != nil {
		log.Errorf("MaskModel failed, getValueModel error:%s", objErr.Error())
		return nil
	}

	return objPtr
}
