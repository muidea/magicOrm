package mysql

import (
	"fmt"
	"log"

	"github.com/muidea/magicOrm/model"
)

// BuildDropSchema  BuildDropSchema
func (s *Builder) BuildDropSchema() (string, error) {
	str := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", s.getTableName(s.modelInfo))
	log.Print(str)

	return str, nil
}

// BuildDropRelationSchema BuildDropRelationSchema
func (s *Builder) BuildDropRelationSchema(fieldName string, relationInfo model.Model) (string, error) {
	str := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", s.GetRelationTableName(fieldName, relationInfo))
	log.Print(str)

	return str, nil
}
