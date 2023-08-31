package mysql

import (
	"fmt"
)

// BuildDropTable  BuildDropSchema
func (s *Builder) BuildDropTable() (string, error) {
	str := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", s.GetTableName())
	//log.Print(str)

	return str, nil
}

// BuildDropRelationTable Build DropRelation Schema
func (s *Builder) BuildDropRelationTable(relationTableName string) (string, error) {
	str := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", relationTableName)
	//log.Print(str)

	return str, nil
}
