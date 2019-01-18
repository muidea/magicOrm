package mysql

import (
	"fmt"
	"log"

	"muidea.com/magicOrm/filter"
)

// BuildBatchQuery BuildBatchQuery
func (s *Builder) BuildBatchQuery(filter filter.Filter) (ret string, err error) {
	ret = fmt.Sprintf("SELECT %s FROM `%s`", s.getFieldQueryNames(s.structInfo), s.getTableName(s.structInfo))
	if filter != nil {
		filterSQL, filterErr := filter.Builder(s.structInfo)
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
