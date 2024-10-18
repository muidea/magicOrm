package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
)

// BuildDropTable  BuildDropSchema
func (s *Builder) BuildDropTable(vModel model.Model) (ret *Result, err *cd.Result) {
	dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", s.buildCodec.ConstructModelTableName(vModel))
	//log.Print(dropSQL)
	if traceSQL() {
		log.Infof("[SQL] drop: %s", dropSQL)
	}

	ret = NewResult(dropSQL, nil)
	return
}

// BuildDropRelationTable Build DropRelation Schema
func (s *Builder) BuildDropRelationTable(vModel model.Model, vField model.Field, rModel model.Model) (ret *Result, err *cd.Result) {
	relationTableName, relationErr := s.buildCodec.ConstructRelationTableName(vModel, vField, rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("BuildDeleteRelation %s failed, s.buildCodec.ConstructRelationTableName error:%s", vField.GetName(), err.Error())
		return
	}

	dropRelationSQL := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", relationTableName)
	//log.Print(dropRelationSQL)
	if traceSQL() {
		log.Infof("[SQL] drop relation: %s", dropRelationSQL)
	}

	ret = NewResult(dropRelationSQL, nil)
	return
}
