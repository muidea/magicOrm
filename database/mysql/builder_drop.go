package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database/context"
	"github.com/muidea/magicOrm/model"
)

// BuildDropTable  BuildDropSchema
func (s *Builder) BuildDropTable() (ret context.BuildResult, err *cd.Result) {
	dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", s.buildContext.BuildHostModelTableName())
	//log.Print(dropSQL)
	if traceSQL() {
		log.Infof("[SQL] drop: %s", dropSQL)
	}

	ret = NewBuildResult(dropSQL, nil)
	return
}

// BuildDropRelationTable Build DropRelation Schema
func (s *Builder) BuildDropRelationTable(vField model.Field, rModel model.Model) (ret context.BuildResult, err *cd.Result) {
	relationTableName, relationErr := s.buildContext.BuildRelationTableName(vField, rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("BuildDeleteRelation %s failed, s.buildContext.BuildRelationTableName error:%s", vField.GetName(), err.Error())
		return
	}

	dropRelationSQL := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", relationTableName)
	//log.Print(dropRelationSQL)
	if traceSQL() {
		log.Infof("[SQL] drop relation: %s", dropRelationSQL)
	}

	ret = NewBuildResult(dropRelationSQL, nil)
	return
}
