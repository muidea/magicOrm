package mysql

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
)

// BuildCreateTable  BuildCreateSchema
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
	lPKType, lErr := getFieldType(lPKField)
	if lErr != nil {
		err = lErr
		return
	}

	rPKField := s.GetPrimaryKeyField(rModel)
	rPKType, rErr := getFieldType(rPKField)
	if rErr != nil {
		err = rErr
		return
	}

	relationTableName := s.GetRelationTableName(field, rModel)
	str := fmt.Sprintf("\t`id` INT NOT NULL AUTO_INCREMENT,\n\t`left` %s NOT NULL,\n\t`right` %s NOT NULL,\n\tPRIMARY KEY (`id`),\n\tINDEX(`left`)", lPKType, rPKType)
	str = fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n%s\n)\n", relationTableName, str)
	//log.Print(str)
	ret = str

	return
}
