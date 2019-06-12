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

	ret = fmt.Sprintf("SELECT %s FROM `%s`", namesVal, s.GetHostTableName(s.modelInfo))
	if filter != nil {
		filterSQL, filterErr := s.buildBatchFilter(filter)
		if filterErr != nil {
			err = filterErr
			return
		}

		if filterSQL != "" {
			ret = fmt.Sprintf("%s WHERE %s", ret, filterSQL)
		}
	}

	//log.Print(ret)
	return
}

func (s *Builder) buildBatchFilter(filter model.Filter) (ret string, err error) {
	filterSQL := ""
	relationFilterSQL := ""
	params := filter.Items()
	fields := s.modelInfo.GetFields()
	for _, field := range fields {
		filterItem, ok := params[field.GetName()]
		if !ok {
			continue
		}

		fType := field.GetType()
		dependModel, dependErr := s.modelProvider.GetTypeModel(fType)
		if dependErr != nil {
			err = dependErr
			return
		}

		if dependModel != nil {
			strVal, strErr := filterItem.FilterStr("right", fType)
			if strErr != nil {
				err = strErr
				return
			}
			if strVal == "" {
				continue
			}

			relationTable := s.GetRelationTableName(field.GetName(), dependModel)
			if relationFilterSQL == "" {
				relationFilterSQL = fmt.Sprintf("SELECT DISTINCT(`left`) `id`  FROM `%s` WHERE %s", relationTable, strVal)
			} else {
				relationFilterSQL = fmt.Sprintf("%s UNION SELECT DISTINCT(`left`) `id` FROM `%s` WHERE %s", relationFilterSQL, relationTable, strVal)
			}

			continue
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

	limit, offset, paging := filter.Pagination()
	if paging {
		ret = fmt.Sprintf("%s LIMIT %d OFFSET %d", ret, limit, offset)
	}

	return
}
