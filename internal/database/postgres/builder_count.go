package postgres

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"

	"log/slog"

	"github.com/muidea/magicOrm/internal/database"
	"github.com/muidea/magicOrm/models"
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

		planWhere, _, planArgs, planErr := database.BuildReadQueryClauseWithCodec(vModel, filter, s.modelProvider, "postgres", len(resultStackPtr.Args()), s.buildCodec)
		if planErr != nil {
			return nil, planErr
		}
		resultStackPtr.PushArgs(planArgs...)
		if planWhere != "" {
			if filterSQL == "" {
				filterSQL = planWhere
			} else {
				filterSQL = "(" + filterSQL + ") AND (" + planWhere + ")"
			}
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
