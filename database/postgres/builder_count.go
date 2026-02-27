package postgres

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/models"
	"log/slog"
)

// BuildCount build count
func (s *Builder) BuildCount(vModel models.Model, filter models.Filter) (ret database.Result, err *cd.Error) {
	resultStackPtr := &ResultStack{}
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM \"%s\"", s.buildCodec.ConstructModelTableName(vModel))
	if filter != nil {
		filterSQL, filterErr := s.buildFilter(vModel, filter, resultStackPtr)
		if filterErr != nil {
			err = filterErr
			slog.Error("BuildCount failed", "operation", "s.buildFilter", "error", err.Error())
			return
		}

		if filterSQL != "" {
			countSQL = fmt.Sprintf("%s WHERE %s", countSQL, filterSQL)
		}
	}

	if traceSQL() {
		slog.Info("[SQL] count", "sql", countSQL)
	}

	resultStackPtr.SetSQL(countSQL)
	ret = resultStackPtr
	return
}
