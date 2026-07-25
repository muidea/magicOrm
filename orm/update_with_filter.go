package orm

import (
	"time"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/metrics"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/validation/errors"
)

// UpdateWithFilter performs a conditional update and returns the exact row
// count reported by the database. It deliberately excludes relation writes:
// a compare-and-set transition must compile to one UPDATE statement.
func (s *impl) UpdateWithFilter(vModel models.Model, filter models.Filter) (rowsAffected int64, err *cd.Error) {
	startTime := time.Now()
	defer func() {
		if ormMetricCollector != nil {
			ormMetricCollector.RecordOperation(string(metrics.OperationUpdate), vModel, time.Since(startTime), cd.ToStdError(err))
		}
	}()

	if err = s.CheckContext(); err != nil {
		return
	}
	if vModel == nil {
		return 0, cd.NewError(cd.IllegalParam, "illegal model value")
	}
	if filter == nil {
		return 0, cd.NewError(cd.IllegalParam, "illegal filter value")
	}
	if !hasAssignedWritableBasicFields(vModel) {
		return 0, cd.NewError(cd.IllegalParam, "no writable fields to update")
	}
	for _, field := range vModel.GetFields() {
		if !models.IsBasicField(field) && models.IsAssignedField(field) {
			return 0, cd.NewError(cd.IllegalParam, "conditional update does not support relation fields")
		}
	}
	if err = s.validateModel(vModel, errors.ScenarioUpdate); err != nil {
		return
	}

	updateResult, buildErr := NewBuilder(s.modelProvider, s.modelCodec).BuildUpdateWithFilter(vModel, filter)
	if buildErr != nil {
		err = buildErr
		return
	}
	return s.executor.Execute(updateResult.SQL(), updateResult.Args()...)
}
