package orm

import (
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/metrics"
	"github.com/muidea/magicOrm/models"
	"log/slog"
)

// BatchQuery batch query
func (s *impl) BatchQuery(filter models.Filter) (ret []models.Model, err *cd.Error) {
	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime)
		if ormMetricCollector != nil {
			// BatchQuery使用filter的mask model
			var model models.Model
			if filter != nil {
				model = filter.MaskModel()
			}
			ormMetricCollector.RecordOperation(string(metrics.OperationBatch), model, duration, err)
		}
	}()

	if filter == nil {
		err = cd.NewError(cd.IllegalParam, "filter is nil")
		slog.Error("BatchQuery: filter is nil")
		return
	}

	responseModel, responseByMask, responseErr := buildQueryResponseModel(nil, filter)
	if responseErr != nil {
		err = responseErr
		slog.Error("BatchQuery buildQueryResponseModel failed", "error", err.Error())
		return
	}

	queryMask, maskErr := buildFullQueryMaskModel(responseModel)
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
