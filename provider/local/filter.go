package local

import (
	"fmt"
	"reflect"

	"log/slog"

	cd "github.com/muidea/magicCommon/def"
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
	bindModel  models.Model
	params     map[string]*filterItem
	maskValue  *ValueImpl
	pageFilter *utils.Pagination
	sortFilter *utils.SortFilter
}

func newFilter(valuePtr *ValueImpl, bindModels ...models.Model) *filter {
	ret := &filter{bindValue: valuePtr, params: map[string]*filterItem{}}
	if len(bindModels) > 0 {
		ret.bindModel = bindModels[0]
	}
	return ret
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
		slog.Error("filter op failed", "error", err.Error())
		return
	}

	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := utils.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		slog.Error("filter op failed", "error", err.Error())
		return
	}
	if models.IsSliceType(qvType) {
		err = cd.NewError(cd.Unexpected, "equal failed, illegal value type")
		slog.Error("filter op failed", "error", err.Error())
		return
	}

	//s.equalFilter = append(s.equalFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: models.EqualOpr, value: NewValue(qv)}
	return
}

func (s *filter) NotEqual(key string, val any) (err *cd.Error) {
	if val == nil {
		err = cd.NewError(cd.IllegalParam, "illegal not equal value")
		slog.Error("filter op failed", "error", err.Error())
		return
	}

	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := utils.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		slog.Error("filter op failed", "error", err.Error())
		return
	}
	if models.IsSliceType(qvType) {
		err = cd.NewError(cd.Unexpected, "NotEqual failed, illegal value type")
		slog.Error("NotEqual failed", "key", key, "error", err.Error())
		return
	}

	//s.notEqualFilter = append(s.notEqualFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: models.NotEqualOpr, value: NewValue(qv)}
	return
}

func (s *filter) Below(key string, val any) (err *cd.Error) {
	if val == nil {
		err = cd.NewError(cd.IllegalParam, "illegal below value")
		slog.Error("filter op failed", "error", err.Error())
		return
	}

	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := utils.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		slog.Error("filter op failed", "error", err.Error())
		return
	}
	if !models.IsBasicType(qvType) {
		err = cd.NewError(cd.Unexpected, "below failed, illegal value type")
		slog.Error("filter op failed", "error", err.Error())
		return
	}

	//s.belowFilter = append(s.belowFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: models.BelowOpr, value: NewValue(qv)}
	return
}

func (s *filter) Above(key string, val any) (err *cd.Error) {
	if val == nil {
		err = cd.NewError(cd.IllegalParam, "illegal above value")
		slog.Error("filter op failed", "error", err.Error())
		return
	}

	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := utils.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		slog.Error("filter op failed", "error", err.Error())
		return
	}
	if !models.IsBasicType(qvType) {
		err = cd.NewError(cd.Unexpected, "above failed, illegal value type")
		slog.Error("filter op failed", "error", err.Error())
		return
	}

	//s.aboveFilter = append(s.aboveFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: models.AboveOpr, value: NewValue(qv)}
	return
}

func (s *filter) In(key string, val any) (err *cd.Error) {
	if val == nil {
		err = cd.NewError(cd.IllegalParam, "illegal in value")
		slog.Error("filter op failed", "error", err.Error())
		return
	}

	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := utils.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		slog.Error("filter op failed", "error", err.Error())
		return
	}
	if !models.IsSliceType(qvType) {
		err = cd.NewError(cd.Unexpected, "in failed, illegal value type")
		slog.Error("filter op failed", "error", err.Error())
		return
	}

	//s.inFilter = append(s.inFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: models.InOpr, value: NewValue(qv)}
	return
}

func (s *filter) NotIn(key string, val any) (err *cd.Error) {
	if val == nil {
		err = cd.NewError(cd.IllegalParam, "illegal not in value")
		slog.Error("filter op failed", "error", err.Error())
		return
	}

	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := utils.GetTypeEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		slog.Error("filter op failed", "error", err.Error())
		return
	}
	if !models.IsSliceType(qvType) {
		err = cd.NewError(cd.Unexpected, "notIn failed, illegal value type")
		slog.Error("filter op failed", "error", err.Error())
		return
	}

	//s.notInFilter = append(s.notInFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: models.NotInOpr, value: NewValue(qv)}
	return
}

func (s *filter) Like(key string, val any) (err *cd.Error) {
	if val == nil {
		err = cd.NewError(cd.IllegalParam, "illegal like value")
		slog.Error("filter op failed", "error", err.Error())
		return
	}

	qv := reflect.Indirect(reflect.ValueOf(val))
	if qv.Kind() != reflect.String {
		err = cd.NewError(cd.Unexpected, "like failed, illegal value type")
		slog.Error("filter op failed", "error", err.Error())
		return
	}

	//s.likeFilter = append(s.likeFilter, &itemValue{name: key, value: newValue(qv)})
	s.params[key] = &filterItem{oprCode: models.LikeOpr, value: NewValue(qv)}
	return
}

func (s *filter) Pagination(pageNum, pageSize int) {
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
		slog.Error("filter op failed", "error", err.Error())
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
		slog.Error("filter Mask failed", "error", objErr.Error())
		return nil
	}

	return objPtr
}

func (s *filter) HasValueMask() bool {
	return s.maskValue != nil
}

func (s *filter) ResponseModel() models.Model {
	if s.bindModel == nil {
		return nil
	}

	return s.bindModel.Copy(models.OriginView)
}

func (s *filter) ExplicitResponseModel() models.Model {
	return s.MaskModel()
}
