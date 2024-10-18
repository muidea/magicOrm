package orm

import (
	"database/sql"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/executor"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

type CountRunner struct {
	baseRunner
}

func NewCountRunner(
	vModel model.Model,
	executor executor.Executor,
	provider provider.Provider,
	modelCodec codec.Codec) *CountRunner {
	return &CountRunner{
		baseRunner: newBaseRunner(vModel, executor, provider, modelCodec, false, 0),
	}
}

func (s *CountRunner) Count(vFilter model.Filter) (ret int64, err *cd.Result) {
	countResult, countErr := s.hBuilder.BuildCount(s.vModel, vFilter)
	if countErr != nil {
		err = countErr
		log.Errorf("Count failed, hBuilder.BuildCount error:%s", err.Error())
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

func (s *impl) Count(vFilter model.Filter) (ret int64, err *cd.Result) {
	if vFilter == nil {
		err = cd.NewError(cd.IllegalParam, "illegal filter value")
		return
	}

	vModel := vFilter.MaskModel()
	countRunner := NewCountRunner(vModel, s.executor, s.modelProvider, s.modelCodec)
	queryVal, queryErr := countRunner.Count(vFilter)
	if queryErr != nil {
		err = queryErr
		log.Errorf("Count failed, countRunner.Count error:%s", err.Error())
		return
	}

	ret = queryVal
	return
}
