package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/models"
	"log/slog"
)

// BuildUpdate  Build Update
func (s *Builder) BuildUpdate(vModel models.Model) (ret database.Result, err *cd.Error) {
	resultStackPtr := &ResultStack{}
	updateStr, updateErr := s.buildFieldUpdateValues(vModel, resultStackPtr)
	if updateErr != nil {
		err = updateErr
		slog.Error("BuildUpdate failed", "value", "s.buildFieldUpdateValues", "error", err.Error())
		return
	}
	filterStr, filterErr := s.buildFieldFilter(vModel.GetPrimaryField(), resultStackPtr)
	if filterErr != nil {
		err = filterErr
		slog.Error("BuildUpdate failed", "value", "s.BuildModelFilter", "error", err.Error())
		return
	}

	updateSQL := fmt.Sprintf("UPDATE `%s` SET %s WHERE %s", s.buildCodec.ConstructModelTableName(vModel), updateStr, filterStr)
	if traceSQL() {
		slog.Info("[SQL] update", "sql", updateSQL)
	}

	resultStackPtr.SetSQL(updateSQL)
	ret = resultStackPtr
	return
}

func (s *Builder) buildFieldUpdateValues(vModel models.Model, resultStackPtr *ResultStack) (ret string, err *cd.Error) {
	str := ""
	for _, field := range vModel.GetFields() {

		if models.IsPrimaryField(field) {
			continue
		}
		if !models.IsBasicField(field) || !models.IsValidField(field) {
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
		encodeVal, encodeErr := s.buildCodec.PackedBasicFieldValue(field, fVal)
		if encodeErr != nil {
			err = encodeErr
			slog.Error("buildFieldUpdateValues failed", "field", field.GetName(), "operation", "encodeFieldValue", "error", err.Error())
			return
		}

		resultStackPtr.PushArgs(encodeVal)
		if str == "" {
			str = fmt.Sprintf("`%s` = ?", field.GetName())
		} else {
			str = fmt.Sprintf("%s,`%s` = ?", str, field.GetName())
		}
	}

	ret = str
	return
}
