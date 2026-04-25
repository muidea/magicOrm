package postgres

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"

	"log/slog"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/models"
)

// BuildDropTable  BuildDropSchema
func (s *Builder) BuildDropTable(vModel models.Model) (ret database.Result, err *cd.Error) {
	dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS \"%s\"", s.buildCodec.ConstructModelTableName(vModel))
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
		slog.Error("BuildDeleteRelation failed", "field", vField.GetName(), "operation", "s.buildCodec.ConstructRelationTableName", "error", err.Error())
		return
	}

	dropRelationSQL := fmt.Sprintf("DROP INDEX IF EXISTS \"%s_index\";\nDROP TABLE IF EXISTS \"%s\"", relationTableName, relationTableName)
	//log.Print(dropRelationSQL)
	if traceSQL() {
		slog.Info("[SQL] drop relation", "sql", dropRelationSQL)
	}

	ret = NewError(dropRelationSQL, nil)
	return
}
