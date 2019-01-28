package mysql

import (
	"fmt"
	"log"

	"muidea.com/magicOrm/filter"
)

// BuildBatchQuery BuildBatchQuery
func (s *Builder) BuildBatchQuery(filter filter.Filter) (ret string, err error) {
	ret = fmt.Sprintf("SELECT %s FROM `%s`", s.getFieldQueryNames(s.modelInfo), s.getTableName(s.modelInfo))
	if filter != nil {
		filterSQL, filterErr := filter.Builder(s.modelInfo)
		if filterErr != nil {
			err = filterErr
			return
		}
		if filterSQL != "" {
			ret = fmt.Sprintf("%s WHERE %s", ret, filterSQL)
		}
	}

	log.Print(ret)
	return
}
