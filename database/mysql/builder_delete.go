package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database/context"
	"github.com/muidea/magicOrm/model"
)

// BuildDelete  BuildDelete
func (s *Builder) BuildDelete() (ret context.BuildResult, err *cd.Result) {
	filterStr, filterErr := s.buildFiledFilter(s.hostModel.GetPrimaryField())
	if filterErr != nil {
		err = filterErr
		log.Errorf("BuildDelete failed, s.BuildModelFilter error:%s", err.Error())
		return
	}

	deleteSQL := fmt.Sprintf("DELETE FROM `%s` WHERE %s", s.buildContext.BuildHostModelTableName(), filterStr)
	if traceSQL() {
		log.Infof("[SQL] delete: %s", deleteSQL)
	}

	ret = NewBuildResult(deleteSQL, nil)
	return
}

// BuildDeleteRelation BuildDeleteRelation
func (s *Builder) BuildDeleteRelation(vField model.Field, rModel model.Model) (delRight, delRelation context.BuildResult, err *cd.Result) {
	leftVal, leftErr := s.buildContext.BuildHostModelValue()
	if leftErr != nil {
		err = leftErr
		log.Errorf("BuildDeleteRelation failed, s.BuildHostModelValue error:%s", err.Error())
		return
	}

	relationTableName, relationErr := s.buildContext.BuildRelationTableName(vField, rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("BuildDeleteRelation %s failed, s.buildContext.BuildRelationTableName error:%s", vField.GetName(), err.Error())
		return
	}

	delRightSQL := fmt.Sprintf("DELETE FROM `%s` WHERE `%s` IN (SELECT `right` FROM `%s` WHERE `left`=%v)",
		s.buildContext.BuildModelTableName(rModel),
		rModel.GetPrimaryField().GetName(),
		relationTableName,
		leftVal)
	delRight = NewBuildResult(delRightSQL, nil)
	//log.Print(delRight)

	delRelationSQL := fmt.Sprintf("DELETE FROM `%s` WHERE `left`=%v", relationTableName, leftVal)
	delRelation = NewBuildResult(delRelationSQL, nil)
	//log.Print(delRelation)
	if traceSQL() {
		log.Infof("[SQL] delete: %s, delete relation: %s", delRight, delRelation)
	}

	return
}
