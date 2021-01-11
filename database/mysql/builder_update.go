package mysql

import (
	"fmt"
	"log"
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

	filterStr, filterErr := s.buildPKFilter()
	if filterErr != nil {
		err = filterErr
		log.Printf("buildPKFilter failed, err:%s", err.Error())
		return
	}

	str = fmt.Sprintf("UPDATE `%s` SET %s WHERE %s", s.getHostTableName(s.modelInfo), str, filterStr)
	//log.Print(str)
	ret = str

	return
}
