package mysql

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
)

// BuildBatchQuery BuildBatchQuery
func (s *Builder) BuildBatchQuery(filter model.Filter) (ret string, err error) {
	namesVal, nameErr := s.getFieldQueryNames(s.modelInfo)
	if nameErr != nil {
		err = nameErr
		return
	}

	ret = fmt.Sprintf("SELECT %s FROM `%s`", namesVal, s.getHostTableName(s.modelInfo))
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
			ret = fmt.Sprintf("%s order by %s", ret, sortVal)
		}

		limit, offset, paging := filter.Pagination()
		if paging {
			ret = fmt.Sprintf("%s LIMIT %d OFFSET %d", ret, limit, offset)
		}
	}

	//log.Print(ret)
	return
}

func (s *Builder) buildFilter(filter model.Filter) (ret string, err error) {
	if filter == nil {
		return
	}

	filterSQL := ""
	fTag := s.modelInfo.GetPrimaryField().GetTag()
	for _, field := range s.modelInfo.GetFields() {
		filterItem := filter.GetFilterItem(field.GetTag().GetName())
		if filterItem == nil {
			continue
		}
		oprValue := filterItem.OprValue()
		oprFunc := filterItem.OprFunc()

		fType := field.GetType()
		valueStr, valueErr := s.buildValue(oprValue, fType)
		if valueErr != nil {
			err = valueErr
			return
		}

		if fType.IsBasic() {
			strVal := oprFunc(field.GetName(), valueStr)
			if filterSQL == "" {
				filterSQL = fmt.Sprintf("%s", strVal)
			} else {
				filterSQL = fmt.Sprintf("%s AND %s", filterSQL, strVal)
			}
			continue
		}

		fieldModel, fieldErr := s.modelProvider.GetTypeModel(fType)
		if fieldErr != nil {
			err = fieldErr
			return
		}

		if fieldModel != nil {
			relationFilterSQL := ""
			strVal := oprFunc("right", valueStr)
			relationTable := s.GetRelationTableName(field.GetName(), fieldModel)
			relationFilterSQL = fmt.Sprintf("SELECT DISTINCT(`left`) `id`  FROM `%s` WHERE %s", relationTable, strVal)
			relationFilterSQL = fmt.Sprintf("`%s` IN (SELECT DISTINCT(`id`) FROM (%s) ids)", fTag.GetName(), relationFilterSQL)

			if filterSQL == "" {
				filterSQL = fmt.Sprintf("%s", relationFilterSQL)
			} else {
				filterSQL = fmt.Sprintf("%s AND %s", filterSQL, relationFilterSQL)
			}
		}
	}

	if filterSQL != "" {
		ret = fmt.Sprintf("%s", filterSQL)
	}

	return
}

func (s *Builder) buildSortFilter(filter model.Sorter) (ret string, err error) {
	if filter == nil {
		return
	}

	for _, val := range s.modelInfo.GetFields() {
		if val.GetTag().GetName() == filter.Name() {
			ret = SortOpr(filter.Name(), filter.AscSort())
			return
		}
	}

	err = fmt.Errorf("illegal sort field name:%s", filter.Name())
	return
}
