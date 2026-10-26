package orm

import (
	"context"
	"database/sql"
	"time"

	cd "github.com/muidea/magicCommon/def"

	"log/slog"

	"github.com/muidea/magicOrm/internal/database"
	"github.com/muidea/magicOrm/internal/database/codec"
	"github.com/muidea/magicOrm/internal/metrics/shared"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

type countRunner struct {
	baseRunner
}

func newCountRunner(
	ctx context.Context,
	vModel models.Model,
	executor database.Executor,
	provider provider.Provider,
	modelCodec codec.Codec) *countRunner {
	return &countRunner{
		baseRunner: newBaseRunner(ctx, vModel, executor, provider, modelCodec, false, 0),
	}
}

func (s *countRunner) Count(vFilter models.Filter) (ret int64, err *cd.Error) {
	if err = s.checkContext(); err != nil {
		return
	}

	countResult, countErr := s.sqlBuilder.BuildCount(s.vModel, vFilter)
	if countErr != nil {
		err = countErr
		slog.Error("countRunner Count BuildCount failed", "error", err.Error())
		return
	}

	_, err = s.executor.Query(countResult.SQL(), false, countResult.Args()...)
	if err != nil {
		return
	}
	defer s.executor.Finish()

	if s.executor.Next() {
		var countVal sql.NullInt64
		err = s.executor.GetField(&countVal)
		if err != nil {
			slog.Error("countRunner Count GetField failed", "error", err.Error())
			return
		}

		ret = countVal.Int64
	}

	return
}

func (s *impl) Count(vFilter models.Filter) (ret int64, err *cd.Error) {
	startTime := time.Now()
	var vModel models.Model

	defer func() {
		duration := time.Since(startTime)
		if ormMetricCollector != nil {
			// Reuse the query model; metrics only consume its identity.
			if vModel == nil && vFilter != nil {
				vModel = vFilter.MaskModel()
			}
			ormMetricCollector.RecordOperation(string(metrics.OperationCount), vModel, duration, cd.ToStdError(err))
		}
	}()

	if err = s.checkContext(); err != nil {
		return
	}

	if vFilter == nil {
		err = cd.NewError(cd.IllegalParam, "illegal filter value")
		return
	}

	vModel = vFilter.MaskModel()
	countRunner := newCountRunner(s.context, vModel, s.executor, s.modelProvider, s.modelCodec)
	queryVal, queryErr := countRunner.Count(vFilter)
	if queryErr != nil {
		err = queryErr
		slog.Error("Count countRunner.Count failed", "error", err.Error())
		return
	}

	ret = queryVal
	return
}
