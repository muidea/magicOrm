package mysql

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
)

// BuildDropTable  BuildDropSchema
func (s *Builder) BuildDropTable() (ret string, err error) {
	str := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", s.GetTableName())
	//log.Print(str)

	ret = str
	return
}

// BuildDropRelationTable Build DropRelation Schema
func (s *Builder) BuildDropRelationTable(field model.Field, rModel model.Model) (ret string, err error) {
	relationTableName := s.GetRelationTableName(field, rModel)
	str := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", relationTableName)
	//log.Print(str)

	ret = str
	return
}
