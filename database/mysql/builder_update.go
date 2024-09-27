package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicCommon/foundation/log"
)

// BuildUpdate  Build Update
func (s *Builder) BuildUpdate() (ret string, err *cd.Result) {
	updateStr, updateErr := s.getFieldUpdateValues()
	if updateErr != nil {
		err = updateErr
		log.Errorf("BuildUpdate failed, s.getFieldUpdateValues error:%s", err.Error())
		return
	}
	filterStr, filterErr := s.common.BuildModelFilter()
	if filterErr != nil {
		err = filterErr
		log.Errorf("BuildUpdate failed, s.BuildModelFilter error:%s", err.Error())
		return
	}

	str := fmt.Sprintf("UPDATE `%s` SET %s WHERE %s", s.common.GetHostTableName(), updateStr, filterStr)
	//log.Print(str)
	if traceSQL() {
		log.Infof("[SQL] update: %s", str)
	}

	ret = str

	return
}

func (s *Builder) getFieldUpdateValues() (ret string, err *cd.Result) {
	str := ""
	for _, field := range s.common.GetHostFields() {
		if field.IsPrimaryKey() {
			continue
		}

		fType := field.GetType()
		fValue := field.GetValue()
		if !fType.IsBasic() || fValue.IsNil() {
			continue
		}

		fStr, fErr := s.common.BuildFieldValue(fType, fValue)
		if fErr != nil {
			err = fErr
			log.Errorf("getFieldUpdateValues failed, BuildFieldValue error:%s", fErr.Error())
			return
		}

		if str == "" {
			str = fmt.Sprintf("`%s` = %v", field.GetName(), fStr)
		} else {
			str = fmt.Sprintf("%s,`%s` = %v", str, field.GetName(), fStr)
		}
	}

	ret = str
	return
}
