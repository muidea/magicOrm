package mysql

import (
	"fmt"

	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
)

// BuildQuery build query sql
func (s *Builder) BuildQuery(filter model.Filter) (ret string, err error) {
	namesVal, nameErr := s.getFieldQueryNames(filter)
	if nameErr != nil {
		err = nameErr
		log.Errorf("BuildQuery failed, s.getFieldQueryNames error:%s", err.Error())
		return
	}

	ret = fmt.Sprintf("SELECT %s FROM `%s`", namesVal, s.GetTableName())
	if filter != nil {
		filterSQL, filterErr := s.buildFilter(filter)
		if filterErr != nil {
			err = filterErr
			log.Errorf("BuildQuery failed, s.buildFilter error:%s", err.Error())
			return
		}

		if filterSQL != "" {
			ret = fmt.Sprintf("%s WHERE %s", ret, filterSQL)
		}

		sortVal, sortErr := s.buildSorter(filter.Sorter())
		if sortErr != nil {
			err = sortErr
			log.Errorf("BuildQuery failed, s.buildSorter error:%s", err.Error())
			return
		}

		if sortVal != "" {
			ret = fmt.Sprintf("%s ORDER BY %s", ret, sortVal)
		}

		limit, offset, paging := filter.Pagination()
		if paging {
			ret = fmt.Sprintf("%s LIMIT %d OFFSET %d", ret, limit, offset)
		}
	}

	//log.Print(ret)
	return
}

// BuildQueryRelation build query relation sql
func (s *Builder) BuildQueryRelation(vField model.Field, rModel model.Model) (ret string, err error) {
	leftVal, leftErr := s.GetModelValue()
	if leftErr != nil {
		err = leftErr
		log.Errorf("BuildQueryRelation failed, s.GetModelValue error:%s", err.Error())
		return
	}

	relationTableName := s.GetRelationTableName(vField, rModel)
	ret = fmt.Sprintf("SELECT `right` FROM `%s` WHERE `left`= %v", relationTableName, leftVal)
	//log.Print(ret)

	return
}

func (s *Builder) buildBasicItem(vField model.Field, filterItem model.FilterItem) (ret string, err error) {
	fType := vField.GetType()
	oprValue := filterItem.OprValue()
	oprFunc := getOprFunc(filterItem)
	valueStr, valueErr := s.EncodeValue(oprValue, fType)
	if valueErr != nil {
		err = valueErr
		log.Errorf("buildBasicItem failed, EncodeValue error:%s", err.Error())
		return
	}

	if fType.IsBasic() {
		ret = oprFunc(vField.GetName(), valueStr)
		return
	}

	err = fmt.Errorf("illegal item type, name:%s", vField.GetName())
	log.Errorf("buildBasicItem failed, error:%s", err.Error())
	return
}

func (s *Builder) buildRelationItem(pkField model.Field, rField model.Field, filterItem model.FilterItem) (ret string, err error) {
	fType := rField.GetType()
	oprValue := filterItem.OprValue()
	oprFunc := getOprFunc(filterItem)
	valueStr, valueErr := s.EncodeValue(oprValue, fType)
	if valueErr != nil {
		err = valueErr
		log.Errorf("buildRelationItem failed, s.EncodeValue error:%s", err.Error())
		return
	}

	fieldModel, fieldErr := s.GetTypeModel(fType)
	if fieldErr != nil {
		err = fieldErr
		log.Errorf("buildRelationItem failed, s.GetTypeModel error:%s", err.Error())
		return
	}

	relationFilterSQL := ""
	strVal := oprFunc("right", valueStr)
	relationTableName := s.GetRelationTableName(rField, fieldModel)
	relationFilterSQL = fmt.Sprintf("SELECT DISTINCT(`left`) `id`  FROM `%s` WHERE %s", relationTableName, strVal)
	relationFilterSQL = fmt.Sprintf("`%s` IN (%s)", pkField.GetName(), relationFilterSQL)
	ret = relationFilterSQL
	return
}

func (s *Builder) buildFilter(filter model.Filter) (ret string, err error) {
	if filter == nil {
		return
	}

	filterSQL := ""
	pkField := s.GetPrimaryKeyField(nil)
	for _, field := range s.GetFields() {
		filterItem := filter.GetFilterItem(field.GetName())
		if filterItem == nil {
			continue
		}

		fType := field.GetType()
		if fType.IsBasic() {
			basicSQL, basicErr := s.buildBasicItem(field, filterItem)
			if basicErr != nil {
				err = basicErr
				log.Errorf("buildFilter failed, s.buildBasicItem error:%s", err.Error())
				return
			}

			if filterSQL == "" {
				filterSQL = fmt.Sprintf("%s", basicSQL)
				continue
			}

			filterSQL = fmt.Sprintf("%s AND %s", filterSQL, basicSQL)
			continue
		}

		relationSQL, relationErr := s.buildRelationItem(pkField, field, filterItem)
		if relationErr != nil {
			err = relationErr
			log.Errorf("buildFilter failed, s.buildRelationItem error:%s", err.Error())
			return
		}

		if filterSQL == "" {
			filterSQL = fmt.Sprintf("%s", relationSQL)
			continue
		}

		filterSQL = fmt.Sprintf("%s AND %s", filterSQL, relationSQL)
	}

	ret = filterSQL
	return
}

func (s *Builder) buildSorter(filter model.Sorter) (ret string, err error) {
	if filter == nil {
		return
	}

	for _, val := range s.GetFields() {
		if val.GetName() == filter.Name() {
			ret = SortOpr(filter.Name(), filter.AscSort())
			return
		}
	}

	err = fmt.Errorf("illegal sort field name:%s", filter.Name())
	log.Errorf("buildSorter failed, err:%s", err.Error())
	return
}

func (s *Builder) getFieldQueryNames(filter model.Filter) (ret string, err error) {
	str := ""
	vModel := filter.MaskModel()
	for _, field := range vModel.GetFields() {
		fType := field.GetType()
		fValue := field.GetValue()
		if !fType.IsBasic() || fValue.IsNil() {
			continue
		}

		if str == "" {
			str = fmt.Sprintf("`%s`", field.GetName())
		} else {
			str = fmt.Sprintf("%s,`%s`", str, field.GetName())
		}
	}

	ret = str
	return
}

func (s *Builder) GetFieldScanDest(vField model.Field) (ret interface{}, err error) {
	return getFieldScanDestPtr(vField)
}
