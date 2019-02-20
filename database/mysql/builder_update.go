package mysql

import (
	"fmt"
	"log"
)

// BuildUpdate  BuildUpdate
func (s *Builder) BuildUpdate() (ret string, err error) {
	str := ""
	for _, val := range s.modelInfo.GetFields() {
		fType := val.GetType()
		fValue := val.GetValue()
		if fValue == nil {
			continue
		}

		if fType.IsPtrType() && fValue.IsNil() {
			continue
		}

		dependModel, dependErr := s.modelProvider.GetTypeModel(fType.GetType())
		if dependErr != nil {
			err = dependErr
			return
		}
		if dependModel != nil {
			continue
		}

		if val != s.modelInfo.GetPrimaryField() {
			fStr, ferr := s.modelProvider.GetValueStr(fType, fValue)
			if ferr != nil {
				err = ferr
				return
			}

			fTag := val.GetTag()
			if str == "" {
				str = fmt.Sprintf("`%s`=%s", fTag.GetName(), fStr)
			} else {
				str = fmt.Sprintf("%s,`%s`=%s", str, fTag.GetName(), fStr)
			}
		}
	}

	pkfVal, pkfErr := s.getStructValue(s.modelInfo)
	if pkfErr != nil {
		err = pkfErr
		return
	}

	pkfTag := s.modelInfo.GetPrimaryField().GetTag()
	str = fmt.Sprintf("UPDATE `%s` SET %s WHERE `%s`=%s", s.getTableName(s.modelInfo), str, pkfTag.GetName(), pkfVal)
	log.Print(str)
	ret = str

	return
}
