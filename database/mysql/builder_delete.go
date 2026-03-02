package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/models"
	"log/slog"
)

// BuildDelete  BuildDelete
func (s *Builder) BuildDelete(vModel models.Model) (ret database.Result, err *cd.Error) {
	resultStackPtr := &ResultStack{}
	filterStr, filterErr := s.buildFieldFilter(vModel.GetPrimaryField(), resultStackPtr)
	if filterErr != nil {
		err = filterErr
		slog.Error("BuildDelete failed", "value", "s.BuildModelFilter", "error", err.Error())
		return
	}

	deleteSQL := fmt.Sprintf("DELETE FROM `%s` WHERE %s", s.buildCodec.ConstructModelTableName(vModel), filterStr)
	if traceSQL() {
		slog.Info("[SQL] delete", "sql", deleteSQL)
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
		slog.Error("BuildDeleteRelation failed", "field", vField.GetName(), "operation", "ConstructRelationTableName", "error", err.Error())
		return
	}

	rModel, rErr := s.modelProvider.GetTypeModel(vField.GetType())
	if rErr != nil {
		err = rErr
		slog.Error("BuildDeleteRelation failed", "field", vField.GetName(), "operation", "GetTypeModel", "error", err.Error())
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
		slog.Info("[SQL] delete host and relation", "host_sql", delHostSQL, "relation_sql", delRelationSQL)
	}

	return
}

// BuildDeleteRelationByRights 仅删除关系表中指定的 (left, right) 行，不删除关联实体
func (s *Builder) BuildDeleteRelationByRights(vModel models.Model, vField models.Field, rightIDs []any) (ret database.Result, err *cd.Error) {
	if len(rightIDs) == 0 {
		return
	}

	hostVal := vModel.GetPrimaryField().GetValue().Get()
	relationTableName, relationErr := s.buildCodec.ConstructRelationTableName(vModel, vField)
	if relationErr != nil {
		err = relationErr
		slog.Error("BuildDeleteRelationByRights failed", "field", vField.GetName(), "operation", "ConstructRelationTableName", "error", err.Error())
		return
	}

	resultStackPtr := &ResultStack{}
	resultStackPtr.PushArgs(hostVal)
	inPlaceholders := "?"
	for i := 1; i < len(rightIDs); i++ {
		inPlaceholders += ",?"
	}
	for _, rightID := range rightIDs {
		resultStackPtr.PushArgs(rightID)
	}
	deleteSQL := fmt.Sprintf("DELETE FROM `%s` WHERE `left`=? AND `right` IN (%s)", relationTableName, inPlaceholders)
	if traceSQL() {
		slog.Info("[SQL] delete relation by rights", "sql", deleteSQL)
	}
	resultStackPtr.SetSQL(deleteSQL)
	ret = resultStackPtr
	return
}
