package orm

import (
	"time"

	"log/slog"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/metrics"
	"github.com/muidea/magicOrm/models"
)

// BatchQuery batch query
func (s *impl) BatchQuery(filter models.Filter) (ret []models.Model, err *cd.Error) {
	startTime := time.Now()
	var metricModel models.Model

	defer func() {
		duration := time.Since(startTime)
		if ormMetricCollector != nil {
			// Preserve error-path labels without rebuilding successful queries.
			if metricModel == nil && filter != nil {
				metricModel = filter.MaskModel()
			}
			ormMetricCollector.RecordOperation(string(metrics.OperationBatch), metricModel, duration, cd.ToStdError(err))
		}
	}()

	if filter == nil {
		err = cd.NewError(cd.IllegalParam, "filter is nil")
		slog.Error("BatchQuery: filter is nil")
		return
	}

	responseModel, responseByMask, responseErr := buildQueryResponseModel(nil, filter)
	metricModel = responseModel
	if responseErr != nil {
		err = responseErr
		slog.Error("BatchQuery buildQueryResponseModel failed", "error", err.Error())
		return
	}

	queryMask, maskErr := buildQueryExecutionModel(responseModel, !responseByMask)
	if maskErr != nil {
		err = maskErr
		slog.Error("BatchQuery buildFullQueryMaskModel failed", "error", err.Error())
		return
	}

	vQueryRunner := NewQueryRunner(s.context, queryMask, responseModel, responseByMask, s.executor, s.modelProvider, s.modelCodec, true, 0)
	queryVal, queryErr := vQueryRunner.Query(filter)
	if queryErr != nil {
		err = queryErr
		slog.Error("BatchQuery QueryRunner.Query failed", "error", err.Error())
		return
	}

	ret = queryVal
	return
}
