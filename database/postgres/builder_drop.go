package postgres

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
)

// BuildDropTable  BuildDropSchema
func (s *Builder) BuildDropTable(vModel model.Model) (ret *ResultStack, err *cd.Error) {
	dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS \"%s\"", s.buildCodec.ConstructModelTableName(vModel))
	//log.Print(dropSQL)
	if traceSQL() {
		log.Infof("[SQL] drop: %s", dropSQL)
	}

	ret = NewError(dropSQL, nil)
	return
}

// BuildDropRelationTable Build DropRelation Schema
func (s *Builder) BuildDropRelationTable(vModel model.Model, vField model.Field) (ret *ResultStack, err *cd.Error) {
	relationTableName, relationErr := s.buildCodec.ConstructRelationTableName(vModel, vField)
	if relationErr != nil {
		err = relationErr
		log.Errorf("BuildDeleteRelation %s failed, s.buildCodec.ConstructRelationTableName error:%s", vField.GetName(), err.Error())
		return
	}

	dropRelationSQL := fmt.Sprintf("DROP TABLE IF EXISTS \"%s\"", relationTableName)
	//log.Print(dropRelationSQL)
	if traceSQL() {
		log.Infof("[SQL] drop relation: %s", dropRelationSQL)
	}

	ret = NewError(dropRelationSQL, nil)
	return
}
