package mysql

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
)

// BuildDropSchema  BuildDropSchema
func (s *Builder) BuildDropSchema() (string, error) {
	str := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", s.getHostTableName(s.modelInfo))
	//log.Print(str)

	return str, nil
}

// BuildDropRelationSchema Build DropRelation Schema
func (s *Builder) BuildDropRelationSchema(fieldName string, relationInfo model.Model) (string, error) {
	str := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", s.GetRelationTableName(fieldName, relationInfo))
	//log.Print(str)

	return str, nil
}
