package orm

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
)

// LockWithFilter locks an exact primary-key row in the current transaction
// without changing its values. A missing or stale row returns zero. Relation
// cleanup remains the responsibility of the ordinary Delete lifecycle.
func LockWithFilter(handler Orm, model models.Model, filter models.Filter) (int64, *cd.Error) {
	capability, ok := handler.(interface {
		LockWithFilter(models.Model, models.Filter) (int64, *cd.Error)
	})
	if !ok || capability == nil {
		return 0, cd.NewError(cd.InvalidOperation, "ORM does not support conditional row locking")
	}
	return capability.LockWithFilter(model, filter)
}

func (s *impl) LockWithFilter(model models.Model, filter models.Filter) (int64, *cd.Error) {
	if s == nil || s.executor == nil {
		return 0, cd.NewError(cd.InvalidOperation, "conditional row lock requires an executor")
	}
	if err := s.checkContext(); err != nil {
		return 0, err
	}
	checked, ok := s.executor.(interface {
		TransactionActive() bool
		RowsError() *cd.Error
	})
	if !ok || !checked.TransactionActive() {
		return 0, cd.NewError(cd.InvalidOperation, "conditional row lock requires an active checked transaction")
	}
	if model == nil || filter == nil || filter.Paginationer() != nil {
		return 0, cd.NewError(cd.IllegalParam, "conditional row lock requires an unpaged model filter")
	}
	if _, readOnly := filter.(models.ReadQueryPlanProvider); readOnly {
		return 0, cd.NewError(cd.IllegalParam, "read query plan cannot authorize row locking")
	}
	if (filter.GetName() != "" && filter.GetName() != model.GetName()) || (filter.GetPkgPath() != "" && filter.GetPkgPath() != model.GetPkgPath()) {
		return 0, cd.NewError(cd.IllegalParam, "conditional row lock filter does not match model")
	}
	primary := model.GetPrimaryField()
	if primary == nil {
		return 0, cd.NewError(cd.IllegalParam, "conditional row lock requires a primary key")
	}
	item := filter.GetFilterItem(primary.GetName())
	if item == nil || item.OprCode() != models.EqualOpr || item.OprValue() == nil || item.OprValue().IsZero() {
		return 0, cd.NewError(cd.IllegalParam, "conditional row lock requires exact primary identity")
	}
	for _, field := range model.GetFields() {
		if filter.GetFilterItem(field.GetName()) != nil && !models.IsMainTableField(model, field) {
			return 0, cd.NewError(cd.IllegalParam, "conditional row lock filter must use main-table fields")
		}
	}
	result, err := newBuilder(s.modelProvider, s.modelCodec).BuildQuery(model, filter)
	if err != nil {
		return 0, err
	}
	if _, err := s.executor.Query(result.SQL()+" FOR UPDATE", false, result.Args()...); err != nil {
		return 0, err
	}
	var count int64
	for s.executor.Next() {
		count++
	}
	s.executor.Finish()
	if err := checked.RowsError(); err != nil {
		return 0, err
	}
	if count > 1 {
		return 0, cd.NewError(cd.DataCorrupted, fmt.Sprintf("primary row lock matched %d rows", count))
	}
	return count, nil
}
