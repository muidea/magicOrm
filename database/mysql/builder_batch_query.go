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
	relationFilterSQL := ""
	for _, field := range s.modelInfo.GetFields() {
		filterItem := filter.GetFilterItem(field.GetName())
		if filterItem == nil {
			continue
		}

		fType := field.GetType()
		if !fType.IsBasic() {
			fieldModel, fieldErr := s.modelProvider.GetTypeModel(fType)
			if fieldErr != nil {
				err = fieldErr
				return
			}

			if fieldModel != nil {
				strVal, strErr := filterItem.FilterStr("right", fType)
				if strErr != nil {
					err = strErr
					return
				}
				if strVal == "" {
					continue
				}

				relationTable := s.GetRelationTableName(field.GetName(), fieldModel)
				if relationFilterSQL == "" {
					relationFilterSQL = fmt.Sprintf("SELECT DISTINCT(`left`) `id`  FROM `%s` WHERE %s", relationTable, strVal)
				} else {
					relationFilterSQL = fmt.Sprintf("%s UNION SELECT DISTINCT(`left`) `id` FROM `%s` WHERE %s", relationFilterSQL, relationTable, strVal)
				}

				continue
			}
		}

		strVal, strErr := filterItem.FilterStr(field.GetName(), fType)
		if strErr != nil {
			err = strErr
			return
		}
		if strVal == "" {
			continue
		}

		if filterSQL == "" {
			filterSQL = fmt.Sprintf("%s", strVal)
		} else {
			filterSQL = fmt.Sprintf("%s AND %s", filterSQL, strVal)
		}
	}

	if relationFilterSQL != "" {
		fTag := s.modelInfo.GetPrimaryField().GetTag()
		relationFilterSQL = fmt.Sprintf("`%s` IN (SELECT DISTINCT(`id`) FROM (%s) ids)", fTag.GetName(), relationFilterSQL)

		if filterSQL == "" {
			ret = fmt.Sprintf("%s", relationFilterSQL)
		} else {
			ret = fmt.Sprintf("%s AND %s", filterSQL, relationFilterSQL)
		}
	} else {
		if filterSQL != "" {
			ret = fmt.Sprintf("%s", filterSQL)
		}
	}

	return
}

func (s *Builder) buildSortFilter(filter model.Sorter) (ret string, err error) {
	if filter == nil {
		return
	}

	for _, val := range s.modelInfo.GetFields() {
		if val.GetName() == filter.Name() {
			ret = filter.SortStr(val.GetTag().GetName())
			return
		}
	}

	err = fmt.Errorf("illegal sort field name:%s", filter.Name())
	return
}
