package postgres

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"

	"log/slog"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/utils"
)

// BuildUpdate  Build Update
func (s *Builder) BuildUpdate(vModel models.Model) (ret database.Result, err *cd.Error) {
	resultStackPtr := &ResultStack{}
	updateStr, updateErr := s.buildFieldUpdateValues(vModel, resultStackPtr)
	if updateErr != nil {
		err = updateErr
		slog.Error("BuildUpdate failed", "operation", "s.buildFieldUpdateValues", "error", err.Error())
		return
	}
	if updateStr == "" {
		err = cd.NewError(cd.IllegalParam, "no writable fields to update")
		slog.Error("BuildUpdate failed", "operation", "s.buildFieldUpdateValues", "error", err.Error())
		return
	}
	filterStr, filterErr := s.buildFieldFilter(vModel.GetPrimaryField(), resultStackPtr)
	if filterErr != nil {
		err = filterErr
		slog.Error("BuildUpdate failed", "operation", "s.BuildModelFilter", "error", err.Error())
		return
	}

	updateSQL := fmt.Sprintf("UPDATE \"%s\" SET %s WHERE %s", s.buildCodec.ConstructModelTableName(vModel), updateStr, filterStr)
	if traceSQL() {
		slog.Info("[SQL] update", "sql", updateSQL)
	}

	resultStackPtr.SetSQL(updateSQL)
	ret = resultStackPtr
	return
}

// BuildUpdateWithFilter builds a conditional single-table update. It is used
// by callers that need compare-and-set semantics and therefore must observe
// the affected row count from the executor.
func (s *Builder) BuildUpdateWithFilter(vModel models.Model, filter models.Filter) (ret database.Result, err *cd.Error) {
	if vModel == nil {
		return nil, cd.NewError(cd.IllegalParam, "model is nil")
	}
	if filter == nil {
		return nil, cd.NewError(cd.IllegalParam, "filter is nil")
	}
	if (filter.GetName() != "" && filter.GetName() != vModel.GetName()) ||
		(filter.GetPkgPath() != "" && filter.GetPkgPath() != vModel.GetPkgPath()) {
		return nil, cd.NewError(cd.IllegalParam, "filter does not match model")
	}

	resultStackPtr := &ResultStack{}
	updateStr, updateErr := s.buildFieldUpdateValues(vModel, resultStackPtr)
	if updateErr != nil {
		return nil, updateErr
	}
	if updateStr == "" {
		return nil, cd.NewError(cd.IllegalParam, "no writable fields to update")
	}

	filterFieldCount := 0
	for _, field := range vModel.GetFields() {
		if filter.GetFilterItem(field.GetName()) == nil {
			continue
		}
		if !models.IsBasicField(field) {
			return nil, cd.NewError(cd.IllegalParam, "conditional update filter must use basic fields only")
		}
		filterFieldCount++
	}
	if filterFieldCount == 0 {
		return nil, cd.NewError(cd.IllegalParam, "conditional update filter is empty")
	}

	filterStr, filterErr := s.buildFilter(vModel, filter, resultStackPtr)
	if filterErr != nil {
		return nil, filterErr
	}
	if filterStr == "" {
		return nil, cd.NewError(cd.IllegalParam, "conditional update filter is empty")
	}

	updateSQL := fmt.Sprintf("UPDATE \"%s\" SET %s WHERE %s", s.buildCodec.ConstructModelTableName(vModel), updateStr, filterStr)
	if traceSQL() {
		slog.Info("[SQL] conditional update", "sql", updateSQL)
	}
	resultStackPtr.SetSQL(updateSQL)
	return resultStackPtr, nil
}

func (s *Builder) buildFieldUpdateValues(vModel models.Model, resultStackPtr *ResultStack) (ret string, err *cd.Error) {
	str := ""
	for _, field := range vModel.GetFields() {

		if models.IsPrimaryField(field) {
			continue
		}
		if !models.IsBasicField(field) || !models.IsAssignedField(field) {
			continue
		}
		// Skip read-only fields in update
		if spec := field.GetSpec(); spec != nil {
			if constraints := spec.GetConstraints(); constraints != nil {
				if constraints.Has(models.KeyReadOnly) {
					continue
				}
			}
		}

		fVal := field.GetValue()
		encodeVal := any(nil)
		if !utils.IsReallyNil(fVal.Get()) || !field.GetType().IsPtrType() {
			var encodeErr *cd.Error
			encodeVal, encodeErr = s.buildCodec.PackedBasicFieldValue(field, fVal)
			if encodeErr != nil {
				err = encodeErr
				slog.Error("buildFieldUpdateValues failed", "field", field.GetName(), "operation", "encodeFieldValue", "error", err.Error())
				return
			}
		}

		resultStackPtr.PushArgs(encodeVal)
		if str == "" {
			str = fmt.Sprintf("\"%s\" = $%d", field.GetName(), len(resultStackPtr.Args()))
		} else {
			str = fmt.Sprintf("%s,\"%s\" = $%d", str, field.GetName(), len(resultStackPtr.Args()))
		}
	}

	ret = str
	return
}
