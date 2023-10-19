package mysql

import (
	"fmt"

	"github.com/muidea/magicCommon/foundation/log"
)

// BuildUpdate  Build Update
func (s *Builder) BuildUpdate() (ret string, err error) {
	updateStr, updateErr := s.getFieldUpdateValues()
	if updateErr != nil {
		err = updateErr
		log.Errorf("BuildUpdate failed, s.getFieldUpdateValues error:%s", err.Error())
		return
	}
	filterStr, filterErr := s.buildModelFilter()
	if filterErr != nil {
		err = filterErr
		log.Errorf("BuildUpdate failed, s.buildModelFilter error:%s", err.Error())
		return
	}

	str := fmt.Sprintf("UPDATE `%s` SET %s WHERE %s", s.GetTableName(), updateStr, filterStr)
	//log.Print(str)
	ret = str

	return
}

func (s *Builder) getFieldUpdateValues() (ret string, err error) {
	str := ""
	for _, field := range s.GetFields() {
		if field.IsPrimaryKey() {
			continue
		}

		fType := field.GetType()
		fValue := field.GetValue()
		if !fType.IsBasic() || fValue.IsNil() {
			continue
		}

		fStr, fErr := s.EncodeValue(fValue, fType)
		if fErr != nil {
			err = fErr
			log.Errorf("getFieldUpdateValues failed, s.EncodeValue error:%s", err.Error())
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
