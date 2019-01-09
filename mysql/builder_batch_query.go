package mysql

import (
	"fmt"
	"log"
)

// BuildBatchQuery BuildBatchQuery
func (s *Builder) BuildBatchQuery() (ret string, err error) {
	ret = fmt.Sprintf("SELECT %s FROM `%s`", s.getFieldQueryNames(s.structInfo), s.getTableName(s.structInfo))

	log.Print(ret)
	return
}
