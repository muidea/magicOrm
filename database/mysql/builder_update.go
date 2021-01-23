package mysql

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
	"log"
)

// BuildUpdate  BuildUpdate
func (s *Builder) BuildUpdate() (ret string, err error) {
	updateStr, updateErr := s.getFieldUpdateValues(s.modelInfo)
	if updateErr != nil {
		err = updateErr
		log.Printf("getFieldUpdateValues failed, err:%s", err.Error())
		return
	}
	filterStr, filterErr := s.buildPKFilter(s.modelInfo)
	if filterErr != nil {
		err = filterErr
		log.Printf("buildPKFilter failed, err:%s", err.Error())
		return
	}

	str := fmt.Sprintf("UPDATE `%s` SET %s WHERE %s", s.getHostTableName(s.modelInfo), updateStr, filterStr)
	//log.Print(str)
	ret = str

	return
}

func (s *Builder) getFieldUpdateValues(info model.Model) (ret string, err error) {
	str := ""
	for _, field := range info.GetFields() {
		if field.IsPrimary() {
			continue
		}

		fType := field.GetType()
		fValue := field.GetValue()
		if !fType.IsBasic() || fValue.IsNil() {
			continue
		}

		fStr, fErr := s.buildValue(fValue, fType)
		if fErr != nil {
			err = fErr
			return
		}

		fTag := field.GetTag()
		if str == "" {
			str = fmt.Sprintf("`%s`=%v", fTag.GetName(), fStr)
		} else {
			str = fmt.Sprintf("%s,`%s`=%v", str, fTag.GetName(), fStr)
		}
	}

	ret = str
	return
}
