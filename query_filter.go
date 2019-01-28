package orm

import (
	"fmt"
	"reflect"

	"muidea.com/magicCommon/foundation/util"
	"muidea.com/magicOrm/builder"
	"muidea.com/magicOrm/model"
	ormutil "muidea.com/magicOrm/util"
)

type filterItem struct {
	name      string
	filterFun func(name string, value model.FieldValue) (string, error)
	value     reflect.Value
}

func (s *filterItem) Verify(fType model.FieldType) (err error) {
	valType := s.value.Type()
	fieldType, fieldErr := model.NewFieldType(valType)
	if fieldErr != nil {
		err = fieldErr
		return
	}
	valDType, _ := fieldType.Depend()
	if valDType != nil {
		fieldType, fieldErr = model.NewFieldType(valDType)
		if fieldErr != nil {
			err = fieldErr
			return
		}
	}

	fdType, _ := fType.Depend()
	if fdType != nil {
		fType, err = model.NewFieldType(fdType)
		if err != nil {
			return
		}
	}

	if fieldType.Value() == fType.Value() {
		return
	}

	err = fmt.Errorf("illegal filter value, name:%s, value type:%s", s.name, valType.String())
	return
}

// queryFilter queryFilter
type queryFilter struct {
	params         map[string]filterItem
	pageFilter     *util.PageFilter
	modelInfoCache model.Cache
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

	s.params[key] = filterItem{name: key, filterFun: equleOpr, value: qv}
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

	s.params[key] = filterItem{name: key, filterFun: notEquleOpr, value: qv}
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

	s.params[key] = filterItem{name: key, filterFun: belowOpr, value: qv}
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

	s.params[key] = filterItem{name: key, filterFun: aboveOpr, value: qv}
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

	s.params[key] = filterItem{name: key, filterFun: inOpr, value: qv}
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

	s.params[key] = filterItem{name: key, filterFun: notInOpr, value: qv}
	return
}

func (s *queryFilter) Like(key string, val interface{}) (err error) {
	qv := reflect.Indirect(reflect.ValueOf(val))
	if qv.Kind() != reflect.String {
		err = fmt.Errorf("illegal value type, type:%s", qv.Type().String())
		return
	}

	s.params[key] = filterItem{name: key, filterFun: likeOpr, value: qv}
	return
}

func (s *queryFilter) PageFilter(filter *util.PageFilter) {
	s.pageFilter = filter
}

func (s *queryFilter) Builder(modelInfo model.Model) (ret string, err error) {
	if modelInfo == nil {
		return
	}

	fields := modelInfo.GetFields()
	for _, field := range *fields {
		fType := field.GetType()
		fDepend, _ := fType.Depend()
		if fDepend != nil {
			continue
		}

		filterItem, ok := s.params[field.GetName()]
		if !ok {
			continue
		}

		verifyErr := filterItem.Verify(fType)
		if verifyErr != nil {
			err = verifyErr
			return
		}

		filterVal := reflect.New(filterItem.value.Type()).Elem()
		filterVal.Set(filterItem.value)

		fValue, fErr := model.NewFieldValue(filterVal.Addr())
		if fErr != nil {
			err = fErr
			return
		}

		strVal, strErr := filterItem.filterFun(field.GetName(), fValue)
		if strErr != nil {
			err = strErr
			return
		}
		if strVal == "" {
			continue
		}

		if ret == "" {
			ret = fmt.Sprintf("%s", strVal)
		} else {
			ret = fmt.Sprintf("%s AND %s", ret, strVal)
		}
	}

	relationSQL, relationErr := s.buildRelation(modelInfo)
	if relationErr != nil {
		err = relationErr
		return
	}
	if relationSQL != "" {
		ret = fmt.Sprintf("%s AND %s", ret, relationSQL)
	}

	return
}

func (s *queryFilter) buildRelation(modelInfo model.Model) (ret string, err error) {
	if modelInfo == nil {
		return
	}

	relationSQL := ""
	builder := builder.NewBuilder(modelInfo)
	fields := modelInfo.GetFields()
	for _, field := range *fields {
		fType := field.GetType()
		fDepend, _ := fType.Depend()
		if fDepend == nil {
			continue
		}

		dependInfo, dependErr := model.GetStructInfo(fDepend, s.modelInfoCache)
		if dependErr != nil {
			err = dependErr
			return
		}

		relationTable := builder.GetRelationTableName(field.GetName(), dependInfo)

		filterItem, ok := s.params[field.GetName()]
		if !ok {
			continue
		}

		filterVal := reflect.New(filterItem.value.Type()).Elem()
		filterVal.Set(filterItem.value)

		fValue, fErr := model.NewFieldValue(filterVal.Addr())
		if fErr != nil {
			err = fErr
			return
		}

		strVal, strErr := filterItem.filterFun("right", fValue)
		if strErr != nil {
			err = strErr
			return
		}
		if strVal == "" {
			continue
		}

		if relationSQL == "" {
			relationSQL = fmt.Sprintf("SELECT DISTINCT(`left`) `id`  FROM `%s` WHERE %s", relationTable, strVal)
		} else {
			relationSQL = fmt.Sprintf("%s UNION SELECT DISTINCT(`left`) `id` FROM `%s` WHERE %s", relationSQL, relationTable, strVal)
		}
	}
	if relationSQL != "" {
		pk := modelInfo.GetPrimaryField()
		fTag := pk.GetTag()
		ret = fmt.Sprintf("`%s` IN (SELECT DISTINCT(`id`) FROM (%s) ids)", fTag.Name(), relationSQL)
	}

	if s.pageFilter != nil {
		limitVal := s.pageFilter.PageSize
		offsetVal := s.pageFilter.PageSize * (s.pageFilter.PageNum - 1)
		if offsetVal < 0 {
			offsetVal = 0
		}
		if limitVal <= 0 {
			limitVal = 100
		}
		ret = fmt.Sprintf("%s LIMIT %d OFFSET %d", ret, limitVal, offsetVal)
	}

	return
}
