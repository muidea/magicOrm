package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
)

// BuildDelete  BuildDelete
func (s *Builder) BuildDelete(vModel model.Model) (ret *ResultStack, err *cd.Result) {
	resultStackPtr := &ResultStack{}
	filterStr, filterErr := s.buildFiledFilter(vModel.GetPrimaryField(), resultStackPtr)
	if filterErr != nil {
		err = filterErr
		log.Errorf("BuildDelete failed, s.BuildModelFilter error:%s", err.Error())
		return
	}

	deleteSQL := fmt.Sprintf("DELETE FROM `%s` WHERE %s", s.buildCodec.ConstructModelTableName(vModel), filterStr)
	if traceSQL() {
		log.Infof("[SQL] delete: %s", deleteSQL)
	}

	resultStackPtr.SetSQL(deleteSQL)
	ret = resultStackPtr
	return
}

// BuildDeleteRelation BuildDeleteRelation
func (s *Builder) BuildDeleteRelation(vModel model.Model, vField model.Field, rModel model.Model) (delHost, delRelation *ResultStack, err *cd.Result) {
	hostVal, hostErr := s.buildCodec.BuildModelValue(vModel)
	if hostErr != nil {
		err = hostErr
		log.Errorf("BuildDeleteRelation failed, s.BuildHostModelValue error:%s", err.Error())
		return
	}

	relationTableName, relationErr := s.buildCodec.ConstructRelationTableName(vModel, vField, rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("BuildDeleteRelation %s failed, s.buildCodec.ConstructRelationTableName error:%s", vField.GetName(), err.Error())
		return
	}

	delHostStackPtr := &ResultStack{}
	delHostSQL := fmt.Sprintf("DELETE FROM `%s` WHERE `%s` IN (SELECT `right` FROM `%s` WHERE `left`=?)",
		s.buildCodec.ConstructModelTableName(rModel),
		rModel.GetPrimaryField().GetName(),
		relationTableName)
	delHostStackPtr.SetSQL(delHostSQL)
	delHostStackPtr.PushArgs(hostVal.Value())
	delHost = delHostStackPtr

	delRelationStackPtr := &ResultStack{}
	delRelationSQL := fmt.Sprintf("DELETE FROM `%s` WHERE `left`=?", relationTableName)
	delRelationStackPtr.SetSQL(delRelationSQL)
	delRelationStackPtr.PushArgs(hostVal.Value())
	delRelation = delRelationStackPtr

	if traceSQL() {
		log.Infof("[SQL] delete host: %s, delete relation: %s", delHost, delRelation)
	}

	return
}
