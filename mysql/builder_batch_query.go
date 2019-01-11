package mysql

import (
	"fmt"
	"log"

	"muidea.com/magicOrm/filter"
)

// BuildBatchQuery BuildBatchQuery
func (s *Builder) BuildBatchQuery(filter filter.Filter) (ret string, err error) {
	ret = fmt.Sprintf("SELECT %s FROM `%s`", s.getFieldQueryNames(s.structInfo), s.getTableName(s.structInfo))

	log.Print(ret)
	return
}
