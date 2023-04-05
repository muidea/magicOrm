package mysql

import (
	"fmt"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/model"
)

// BuildQuery build query sql
func (s *Builder) BuildQuery(filter model.Filter) (ret string, err error) {
	namesVal, nameErr := s.getFieldQueryNames(s.entityModel)
	if nameErr != nil {
		err = nameErr
		return
	}

	ret = fmt.Sprintf("SELECT %s FROM `%s`", namesVal, s.GetTableName())
	if filter != nil {
		filterSQL, filterErr := s.buildFilter(filter)
		if filterErr != nil {
			err = filterErr
			return
		}

		if filterSQL != "" {
			ret = fmt.Sprintf("%s WHERE %s", ret, filterSQL)
		}

		sortVal, sortErr := s.buildSortFilter(filter.Sorter())
		if sortErr != nil {
			err = sortErr
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
func (s *Builder) BuildQueryRelation(field model.Field, relationInfo model.Model) (ret string, err error) {
	pkfStr, pkfErr := s.getModelValue()
	if pkfErr != nil {
		err = pkfErr
		return
	}

	relationSchema := s.GetRelationTableName(field, relationInfo)
	ret = fmt.Sprintf("SELECT `right` FROM `%s` WHERE `left`= %v", relationSchema, pkfStr)
	//log.Print(ret)

	return
}

func (s *Builder) buildBasicItem(field model.Field, filterItem model.FilterItem) (ret string, err error) {
	fType := field.GetType()
	oprValue := filterItem.OprValue()
	oprFunc := filterItem.OprFunc()
	valueStr, valueErr := s.encodeValue(oprValue, fType)
	if valueErr != nil {
		err = valueErr
		log.Errorf("encodeValue failed, field name:%s, err:%s", field.GetName(), err.Error())
		return
	}

	if fType.IsBasic() {
		ret = oprFunc(field.GetName(), valueStr)
		return
	}

	err = fmt.Errorf("illegal item type, name:%s", field.GetName())
	return
}

func (s *Builder) buildRelationItem(pkField model.Field, field model.Field, filterItem model.FilterItem) (ret string, err error) {
	fType := field.GetType()
	oprValue := filterItem.OprValue()
	oprFunc := filterItem.OprFunc()
	valueStr, valueErr := s.encodeValue(oprValue, fType)
	if valueErr != nil {
		err = valueErr
		log.Errorf("encodeValue failed, field name:%s, err:%s", field.GetName(), err.Error())
		return
	}

	fieldModel, fieldErr := s.modelProvider.GetTypeModel(fType)
	if fieldErr != nil {
		err = fieldErr
		log.Errorf("GetTypeModel failed, field name:%s, err:%s", field.GetName(), err.Error())
		return
	}

	relationFilterSQL := ""
	strVal := oprFunc("right", valueStr)
	relationTable := s.GetRelationTableName(field, fieldModel)
	relationFilterSQL = fmt.Sprintf("SELECT DISTINCT(`left`) `id`  FROM `%s` WHERE %s", relationTable, strVal)
	relationFilterSQL = fmt.Sprintf("`%s` IN (SELECT DISTINCT(`id`) FROM (%s) ids)", pkField.GetTag().GetName(), relationFilterSQL)
	ret = relationFilterSQL
	return
}

func (s *Builder) buildFilter(filter model.Filter) (ret string, err error) {
	if filter == nil {
		return
	}

	filterSQL := ""
	pkField := s.entityModel.GetPrimaryField()
	for _, field := range s.entityModel.GetFields() {
		filterItem := filter.GetFilterItem(field.GetTag().GetName())
		if filterItem == nil {
			continue
		}

		fType := field.GetType()
		if fType.IsBasic() {
			basicSQL, basicErr := s.buildBasicItem(field, filterItem)
			if basicErr != nil {
				err = basicErr
				return
			}

			if filterSQL == "" {
				filterSQL = fmt.Sprintf("%s", basicSQL)
			} else {
				filterSQL = fmt.Sprintf("%s AND %s", filterSQL, basicSQL)
			}
			continue
		}

		relationSQL, relationErr := s.buildRelationItem(pkField, field, filterItem)
		if relationErr != nil {
			err = relationErr
			return
		}
		if filterSQL == "" {
			filterSQL = fmt.Sprintf("%s", relationSQL)
		} else {
			filterSQL = fmt.Sprintf("%s AND %s", filterSQL, relationSQL)
		}
	}

	ret = filterSQL
	return
}

func (s *Builder) buildSortFilter(filter model.Sorter) (ret string, err error) {
	if filter == nil {
		return
	}

	for _, val := range s.entityModel.GetFields() {
		if val.GetTag().GetName() == filter.Name() {
			ret = SortOpr(filter.Name(), filter.AscSort())
			return
		}
	}

	err = fmt.Errorf("illegal sort field name:%s", filter.Name())
	log.Warnf("buildSortFilter failed, err:%s", err.Error())
	return
}

func (s *Builder) getFieldQueryNames(info model.Model) (ret string, err error) {
	str := ""
	for _, field := range info.GetFields() {
		fType := field.GetType()
		if !fType.IsBasic() {
			continue
		}

		fTag := field.GetTag()
		if str == "" {
			str = fmt.Sprintf("`%s`", fTag.GetName())
		} else {
			str = fmt.Sprintf("%s,`%s`", str, fTag.GetName())
		}
	}

	ret = str
	return
}
