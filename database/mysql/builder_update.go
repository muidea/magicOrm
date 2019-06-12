package mysql

import (
	"fmt"
)

// BuildUpdate  BuildUpdate
func (s *Builder) BuildUpdate() (ret string, err error) {
	str := ""
	for _, val := range s.modelInfo.GetFields() {
		if val.IsPrimary() {
			continue
		}

		fStr, isNil, fErr := s.getFieldValue(val)
		if fErr != nil {
			err = fErr
			return
		}
		if isNil {
			continue
		}

		fTag := val.GetTag()
		if str == "" {
			str = fmt.Sprintf("`%s`=%s", fTag.GetName(), fStr)
		} else {
			str = fmt.Sprintf("%s,`%s`=%s", str, fTag.GetName(), fStr)
		}
	}

	pkfStr, pkfErr := s.getStructValue(s.modelInfo)
	if pkfErr != nil {
		err = pkfErr
		return
	}

	pkfTag := s.modelInfo.GetPrimaryField().GetTag()
	str = fmt.Sprintf("UPDATE `%s` SET %s WHERE `%s`=%s", s.GetHostTableName(s.modelInfo), str, pkfTag.GetName(), pkfStr)
	//log.Print(str)
	ret = str

	return
}
