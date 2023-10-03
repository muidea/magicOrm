package mysql

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
)

func (s *Builder) BuildCreateTable() (ret string, err error) {
	str := ""
	for _, val := range s.GetFields() {
		fType := val.GetType()
		if !fType.IsBasic() {
			continue
		}

		infoVal, infoErr := declareFieldInfo(val)
		if infoErr != nil {
			err = infoErr
			return
		}

		if str == "" {
			str = fmt.Sprintf("\t%s", infoVal)
		} else {
			str = fmt.Sprintf("%s,\n\t%s", str, infoVal)
		}
	}

	pkField := s.GetPrimaryKeyField(nil)
	if pkField != nil {
		str = fmt.Sprintf("%s,\n\tPRIMARY KEY (`%s`)", str, pkField.GetName())
	}

	str = fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n%s\n)\n", s.GetTableName(), str)
	//log.Print(str)

	ret = str
	return
}

// BuildCreateRelationTable Build CreateRelation Schema
func (s *Builder) BuildCreateRelationTable(field model.Field, rModel model.Model) (ret string, err error) {
	lPKField := s.GetPrimaryKeyField(nil)
	lPKType, lPKErr := getTypeDeclare(lPKField.GetType())
	if lPKErr != nil {
		err = lPKErr
		return
	}

	rPKField := s.GetPrimaryKeyField(rModel)
	rPKType, rPKErr := getTypeDeclare(rPKField.GetType())
	if rPKErr != nil {
		err = rPKErr
		return
	}

	relationTableName := s.GetRelationTableName(field, rModel)
	str := fmt.Sprintf("\t`id` INT NOT NULL AUTO_INCREMENT,\n\t`left` %s NOT NULL,\n\t`right` %s NOT NULL,\n\tPRIMARY KEY (`id`),\n\tINDEX(`left`)", lPKType, rPKType)
	str = fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n%s\n)\n", relationTableName, str)
	//log.Print(str)

	ret = str
	return
}
