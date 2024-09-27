package mysql

import (
	"fmt"
	"github.com/muidea/magicCommon/foundation/log"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/model"
)

// BuildDropTable  BuildDropSchema
func (s *Builder) BuildDropTable() (ret string, err *cd.Result) {
	str := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", s.common.GetTableName())
	//log.Print(str)
	if traceSQL() {
		log.Infof("[SQL] drop: %s", str)
	}

	ret = str
	return
}

// BuildDropRelationTable Build DropRelation Schema
func (s *Builder) BuildDropRelationTable(field model.Field, rModel model.Model) (ret string, err *cd.Result) {
	relationTableName := s.common.GetRelationTableName(field, rModel)
	str := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", relationTableName)
	//log.Print(str)
	if traceSQL() {
		log.Infof("[SQL] drop relation: %s", str)
	}

	ret = str
	return
}
