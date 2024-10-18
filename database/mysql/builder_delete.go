package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
)

// BuildDelete  BuildDelete
func (s *Builder) BuildDelete(vModel model.Model) (ret *Result, err *cd.Result) {
	filterStr, filterErr := s.buildFiledFilter(vModel.GetPrimaryField())
	if filterErr != nil {
		err = filterErr
		log.Errorf("BuildDelete failed, s.BuildModelFilter error:%s", err.Error())
		return
	}

	deleteSQL := fmt.Sprintf("DELETE FROM `%s` WHERE %s", s.buildCodec.ConstructModelTableName(vModel), filterStr)
	if traceSQL() {
		log.Infof("[SQL] delete: %s", deleteSQL)
	}

	ret = NewResult(deleteSQL, nil)
	return
}

// BuildDeleteRelation BuildDeleteRelation
func (s *Builder) BuildDeleteRelation(vModel model.Model, vField model.Field, rModel model.Model) (delRight, delRelation *Result, err *cd.Result) {
	leftVal, leftErr := s.buildCodec.BuildModelValue(vModel)
	if leftErr != nil {
		err = leftErr
		log.Errorf("BuildDeleteRelation failed, s.BuildHostModelValue error:%s", err.Error())
		return
	}

	relationTableName, relationErr := s.buildCodec.ConstructRelationTableName(vModel, vField, rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("BuildDeleteRelation %s failed, s.buildCodec.ConstructRelationTableName error:%s", vField.GetName(), err.Error())
		return
	}

	delRightSQL := fmt.Sprintf("DELETE FROM `%s` WHERE `%s` IN (SELECT `right` FROM `%s` WHERE `left`=%v)",
		s.buildCodec.ConstructModelTableName(rModel),
		rModel.GetPrimaryField().GetName(),
		relationTableName,
		leftVal)
	delRight = NewResult(delRightSQL, nil)
	//log.Print(delRight)

	delRelationSQL := fmt.Sprintf("DELETE FROM `%s` WHERE `left`=%v", relationTableName, leftVal)
	delRelation = NewResult(delRelationSQL, nil)
	//log.Print(delRelation)
	if traceSQL() {
		log.Infof("[SQL] delete: %s, delete relation: %s", delRight, delRelation)
	}

	return
}
