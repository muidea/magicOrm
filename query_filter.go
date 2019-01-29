package orm

import (
	"fmt"
	"reflect"

	"muidea.com/magicCommon/foundation/util"
	"muidea.com/magicOrm/local"
	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/mysql"
	ormutil "muidea.com/magicOrm/util"
)

type filterItem struct {
	filterFun func(name string, value model.FieldValue) (string, error)
	value     reflect.Value
}

func (s *filterItem) Verify(fType model.FieldType) (err error) {
	valType := s.value.Type()
	fieldType, fieldErr := local.NewFieldType(valType)
	if fieldErr != nil {
		err = fieldErr
		return
	}
	valDType := fieldType.Depend()
	if valDType != nil {
		fieldType, fieldErr = local.NewFieldType(valDType.Type())
		if fieldErr != nil {
			err = fieldErr
			return
		}
	}

	fdType := fType.Depend()
	if fdType != nil {
		fType, err = local.NewFieldType(fdType.Type())
		if err != nil {
			return
		}
	}

	if fieldType.Value() == fType.Value() {
		return
	}

	err = fmt.Errorf("illegal filter value, value type:%s", valType.String())
	return
}

func (s *filterItem) FilterStr(name string) (ret string, err error) {
	filterVal := reflect.New(s.value.Type()).Elem()
	filterVal.Set(s.value)

	fValue, fErr := local.NewFieldValue(filterVal.Addr())
	if fErr != nil {
		err = fErr
		return
	}

	strVal, strErr := s.filterFun(name, fValue)
	if strErr != nil {
		err = strErr
		return
	}

	ret = strVal
	return
}

// queryFilter queryFilter
type queryFilter struct {
	params     map[string]model.FilterItem
	pageFilter *util.PageFilter
}

func (s *queryFilter) Equle(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ormutil.GetTypeValueEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if ormutil.IsSliceType(qvType) {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	s.params[key] = &filterItem{filterFun: mysql.EquleOpr, value: qv}
	return
}

func (s *queryFilter) NotEqule(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ormutil.GetTypeValueEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if ormutil.IsSliceType(qvType) {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	s.params[key] = &filterItem{filterFun: mysql.NotEquleOpr, value: qv}
	return
}

func (s *queryFilter) Below(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ormutil.GetTypeValueEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if !ormutil.IsBasicType(qvType) {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	s.params[key] = &filterItem{filterFun: mysql.BelowOpr, value: qv}
	return
}

func (s *queryFilter) Above(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ormutil.GetTypeValueEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if !ormutil.IsBasicType(qvType) {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	s.params[key] = &filterItem{filterFun: mysql.AboveOpr, value: qv}
	return
}

func (s *queryFilter) In(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ormutil.GetTypeValueEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if !ormutil.IsSliceType(qvType) {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	s.params[key] = &filterItem{filterFun: mysql.InOpr, value: qv}
	return
}

func (s *queryFilter) NotIn(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	qvType, qvErr := ormutil.GetTypeValueEnum(qv.Type())
	if qvErr != nil {
		err = qvErr
		return
	}
	if !ormutil.IsSliceType(qvType) {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	s.params[key] = &filterItem{filterFun: mysql.NotInOpr, value: qv}
	return
}

func (s *queryFilter) Like(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	if qv.Kind() != reflect.String {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	s.params[key] = &filterItem{filterFun: mysql.LikeOpr, value: qv}
	return
}

func (s *queryFilter) PageFilter(filter *util.PageFilter) {
	s.pageFilter = filter
}

func (s *queryFilter) Items() map[string]model.FilterItem {
	return s.params
}

func (s *queryFilter) Pagination() (limit, offset int, paging bool) {
	paging = false
	if s.pageFilter == nil {
		return
	}

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
