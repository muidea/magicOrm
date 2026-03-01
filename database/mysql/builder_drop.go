package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/models"
	"log/slog"
)

// BuildDropTable  BuildDropSchema
func (s *Builder) BuildDropTable(vModel models.Model) (ret database.Result, err *cd.Error) {
	dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", s.buildCodec.ConstructModelTableName(vModel))
	//log.Print(dropSQL)
	if traceSQL() {
		slog.Info("[SQL] drop", "sql", dropSQL)
	}

	ret = NewError(dropSQL, nil)
	return
}

// BuildDropRelationTable Build DropRelation Schema
func (s *Builder) BuildDropRelationTable(vModel models.Model, vField models.Field) (ret database.Result, err *cd.Error) {
	relationTableName, relationErr := s.buildCodec.ConstructRelationTableName(vModel, vField)
	if relationErr != nil {
		err = relationErr
		slog.Error("BuildDeleteRelation failed", "field", vField.GetName(), "operation", "ConstructRelationTableName", "error", err.Error())
		return
	}

	dropRelationSQL := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", relationTableName)
	//log.Print(dropRelationSQL)
	if traceSQL() {
		slog.Info("[SQL] drop relation", "sql", dropRelationSQL)
	}

	ret = NewError(dropRelationSQL, nil)
	return
}
