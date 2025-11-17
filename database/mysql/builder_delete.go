package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/models"
)

// BuildDelete  BuildDelete
func (s *Builder) BuildDelete(vModel models.Model) (ret database.Result, err *cd.Error) {
	resultStackPtr := &ResultStack{}
	filterStr, filterErr := s.buildFieldFilter(vModel.GetPrimaryField(), resultStackPtr)
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
func (s *Builder) BuildDeleteRelation(vModel models.Model, vField models.Field) (delHost, delRelation database.Result, err *cd.Error) {
	hostVal := vModel.GetPrimaryField().GetValue().Get()
	relationTableName, relationErr := s.buildCodec.ConstructRelationTableName(vModel, vField)
	if relationErr != nil {
		err = relationErr
		log.Errorf("BuildDeleteRelation %s failed, s.buildCodec.ConstructRelationTableName error:%s", vField.GetName(), err.Error())
		return
	}

	rModel, rErr := s.modelProvider.GetTypeModel(vField.GetType())
	if rErr != nil {
		err = rErr
		log.Errorf("BuildDeleteRelation %s failed, s.modelProvider.GetTypeModel error:%s", vField.GetName(), err.Error())
		return
	}

	delHostStackPtr := &ResultStack{}
	delHostSQL := fmt.Sprintf("DELETE FROM `%s` WHERE `%s` IN (SELECT `right` FROM `%s` WHERE `left`=?)",
		s.buildCodec.ConstructModelTableName(vField.GetType()),
		rModel.GetPrimaryField().GetName(),
		relationTableName)
	delHostStackPtr.SetSQL(delHostSQL)
	delHostStackPtr.PushArgs(hostVal)
	delHost = delHostStackPtr

	delRelationStackPtr := &ResultStack{}
	delRelationSQL := fmt.Sprintf("DELETE FROM `%s` WHERE `left`=?", relationTableName)
	delRelationStackPtr.SetSQL(delRelationSQL)
	delRelationStackPtr.PushArgs(hostVal)
	delRelation = delRelationStackPtr

	if traceSQL() {
		log.Infof("[SQL] delete host: %s, delete relation: %s", delHostSQL, delRelationSQL)
	}

	return
}
