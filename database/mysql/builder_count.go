package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"

	"log/slog"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/models"
)

// BuildCount build count
func (s *Builder) BuildCount(vModel models.Model, filter models.Filter) (ret database.Result, err *cd.Error) {
	resultStackPtr := &ResultStack{}
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM `%s`", s.buildCodec.ConstructModelTableName(vModel))
	if filter != nil {
		filterSQL, filterErr := s.buildFilter(vModel, filter, resultStackPtr)
		if filterErr != nil {
			err = filterErr
			slog.Error("BuildCount failed", "value", "s.buildFilter", "error", err.Error())
			return
		}

		planWhere, _, planArgs, planErr := database.BuildReadQueryClause(vModel, filter, s.modelProvider, "mysql", len(resultStackPtr.Args()))
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
		slog.Info("message")
	}

	resultStackPtr.SetSQL(countSQL)
	ret = resultStackPtr
	return
}
