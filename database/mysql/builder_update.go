package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/models"
)

// BuildUpdate  Build Update
func (s *Builder) BuildUpdate(vModel models.Model) (ret database.Result, err *cd.Error) {
	resultStackPtr := &ResultStack{}
	updateStr, updateErr := s.buildFieldUpdateValues(vModel, resultStackPtr)
	if updateErr != nil {
		err = updateErr
		log.Errorf("BuildUpdate failed, s.buildFieldUpdateValues error:%s", err.Error())
		return
	}
	filterStr, filterErr := s.buildFieldFilter(vModel.GetPrimaryField(), resultStackPtr)
	if filterErr != nil {
		err = filterErr
		log.Errorf("BuildUpdate failed, s.BuildModelFilter error:%s", err.Error())
		return
	}

	updateSQL := fmt.Sprintf("UPDATE `%s` SET %s WHERE %s", s.buildCodec.ConstructModelTableName(vModel), updateStr, filterStr)
	if traceSQL() {
		log.Infof("[SQL] update: %s", updateSQL)
	}

	resultStackPtr.SetSQL(updateSQL)
	ret = resultStackPtr
	return
}

func (s *Builder) buildFieldUpdateValues(vModel models.Model, resultStackPtr *ResultStack) (ret string, err *cd.Error) {
	str := ""
	for _, field := range vModel.GetFields() {
		// 检查 ro 和 imm 约束，这些字段在更新时应该被忽略
		fSpec := field.GetSpec()
		constraints := fSpec.GetConstraints()
		if constraints != nil {
			if constraints.Has(models.KeyReadOnly) {
				continue
			}
		}

		if models.IsPrimaryField(field) {
			continue
		}
		if !models.IsBasicField(field) || !models.IsValidField(field) {
			continue
		}

		fVal := field.GetValue()
		encodeVal, encodeErr := s.buildCodec.PackedBasicFieldValue(field, fVal)
		if encodeErr != nil {
			err = encodeErr
			log.Errorf("buildFieldUpdateValues %s failed, encodeFieldValue error:%s", field.GetName(), err.Error())
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
