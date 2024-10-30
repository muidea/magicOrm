package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
)

// BuildUpdate  Build Update
func (s *Builder) BuildUpdate(vModel model.Model) (ret *ResultStack, err *cd.Result) {
	resultStackPtr := &ResultStack{}
	updateStr, updateErr := s.buildFieldUpdateValues(vModel, resultStackPtr)
	if updateErr != nil {
		err = updateErr
		log.Errorf("BuildUpdate failed, s.buildFieldUpdateValues error:%s", err.Error())
		return
	}
	filterStr, filterErr := s.buildFiledFilter(vModel.GetPrimaryField(), resultStackPtr)
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

func (s *Builder) buildFieldUpdateValues(vModel model.Model, resultStackPtr *ResultStack) (ret string, err *cd.Result) {
	str := ""
	for _, field := range vModel.GetFields() {
		if field.IsPrimaryKey() {
			continue
		}

		fType := field.GetType()
		fValue := field.GetValue()
		if !fType.IsBasic() || !fValue.IsValid() {
			continue
		}

		fVal, fErr := s.buildCodec.BuildFieldValue(field)
		if fErr != nil {
			err = fErr
			log.Errorf("buildFieldUpdateValues failed, BuildFieldValue error:%s", fErr.Error())
			return
		}

		resultStackPtr.PushArgs(fVal.Value())
		if str == "" {
			str = fmt.Sprintf("`%s` = ?", field.GetName())
		} else {
			str = fmt.Sprintf("%s,`%s` = ?", str, field.GetName())
		}
	}

	ret = str
	return
}
