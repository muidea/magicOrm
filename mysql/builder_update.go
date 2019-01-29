package mysql

import (
	"fmt"
	"log"
)

// BuildUpdate  BuildUpdate
func (s *Builder) BuildUpdate() (ret string, err error) {
	str := ""
	for _, val := range *s.modelInfo.GetFields() {
		fType := val.GetType()
		fValue := val.GetValue()
		fTag := val.GetTag()

		if fValue == nil {
			continue
		}

		if fType.IsPtr() && fValue.IsNil() {
			continue
		}

		dependType := fType.Depend()
		if dependType != nil {
			continue
		}

		if val != s.modelInfo.GetPrimaryField() {
			fStr, ferr := fValue.ValueStr()
			if ferr != nil {
				err = ferr
				break
			}
			if str == "" {
				str = fmt.Sprintf("`%s`=%s", fTag.Name(), fStr)
			} else {
				str = fmt.Sprintf("%s,`%s`=%s", str, fTag.Name(), fStr)
			}
		}
	}

	if err != nil {
		return
	}

	pkfValue := s.modelInfo.GetPrimaryField().GetValue()
	pkfTag := s.modelInfo.GetPrimaryField().GetTag()
	pkfStr, pkferr := pkfValue.ValueStr()
	if pkferr == nil {
		str = fmt.Sprintf("UPDATE `%s` SET %s WHERE `%s`=%s", s.getTableName(s.modelInfo), str, pkfTag.Name(), pkfStr)
		log.Print(str)
	}

	ret = str
	err = pkferr

	return
}
