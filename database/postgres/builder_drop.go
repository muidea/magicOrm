package postgres

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/builder"
	"github.com/muidea/magicOrm/model"
)

// BuildDropTable  BuildDropSchema
func (s *Builder) BuildDropTable(vModel model.Model) (ret builder.Result, err *cd.Error) {
	dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS \"%s\"", s.buildCodec.ConstructModelTableName(vModel))
	//log.Print(dropSQL)
	if traceSQL() {
		log.Infof("[SQL] drop: %s", dropSQL)
	}

	ret = NewError(dropSQL, nil)
	return
}

// BuildDropRelationTable Build DropRelation Schema
func (s *Builder) BuildDropRelationTable(vModel model.Model, vField model.Field) (ret builder.Result, err *cd.Error) {
	relationTableName, relationErr := s.buildCodec.ConstructRelationTableName(vModel, vField)
	if relationErr != nil {
		err = relationErr
		log.Errorf("BuildDeleteRelation %s failed, s.buildCodec.ConstructRelationTableName error:%s", vField.GetName(), err.Error())
		return
	}

	dropRelationSQL := fmt.Sprintf("DROP INDEX IF EXISTS \"%s_index\";\nDROP TABLE IF EXISTS \"%s\"", relationTableName, relationTableName)
	//log.Print(dropRelationSQL)
	if traceSQL() {
		log.Infof("[SQL] drop relation: %s", dropRelationSQL)
	}

	ret = NewError(dropRelationSQL, nil)
	return
}
