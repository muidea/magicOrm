package mysql

import (
	"fmt"
)

// BuildDropSchema  BuildDropSchema
func (s *Builder) BuildDropSchema() (string, error) {
	str := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", s.GetTableName())
	//log.Print(str)

	return str, nil
}

// BuildDropRelationSchema Build DropRelation Schema
func (s *Builder) BuildDropRelationSchema(relationSchema string) (string, error) {
	str := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", relationSchema)
	//log.Print(str)

	return str, nil
}
