package mysql

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
)

// BuildDropTable  BuildDropSchema
func (s *Builder) BuildDropTable() (string, error) {
	str := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", s.GetTableName())
	//log.Print(str)

	return str, nil
}

// BuildDropRelationTable Build DropRelation Schema
func (s *Builder) BuildDropRelationTable(field model.Field, rModel model.Model) (string, error) {
	relationTableName := s.GetRelationTableName(field, rModel)
	str := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", relationTableName)
	//log.Print(str)

	return str, nil
}
