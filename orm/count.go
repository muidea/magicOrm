package orm

import (
	"context"
	"database/sql"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

type CountRunner struct {
	baseRunner
}

func NewCountRunner(
	ctx context.Context,
	vModel models.Model,
	executor database.Executor,
	provider provider.Provider,
	modelCodec codec.Codec) *CountRunner {
	return &CountRunner{
		baseRunner: newBaseRunner(ctx, vModel, executor, provider, modelCodec, false, 0),
	}
}

func (s *CountRunner) Count(vFilter models.Filter) (ret int64, err *cd.Error) {
	countResult, countErr := s.sqlBuilder.BuildCount(s.vModel, vFilter)
	if countErr != nil {
		err = countErr
		log.Errorf("Count failed, sqlBuilder.BuildCount error:%s", err.Error())
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
			log.Errorf("Count failed, s.executor.GetField error:%s", err.Error())
			return
		}

		ret = countVal.Int64
	}

	return
}

func (s *impl) Count(vFilter models.Filter) (ret int64, err *cd.Error) {
	if vFilter == nil {
		err = cd.NewError(cd.IllegalParam, "illegal filter value")
		return
	}

	vModel := vFilter.MaskModel()
	countRunner := NewCountRunner(s.context, vModel, s.executor, s.modelProvider, s.modelCodec)
	queryVal, queryErr := countRunner.Count(vFilter)
	if queryErr != nil {
		err = queryErr
		log.Errorf("Count failed, countRunner.Count error:%s", err.Error())
		return
	}

	ret = queryVal
	return
}
