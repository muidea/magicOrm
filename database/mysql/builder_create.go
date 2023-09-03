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

	pkField := s.GetPrimaryKeyField()
	if pkField != nil {
		str = fmt.Sprintf("%s,\n\tPRIMARY KEY (`%s`)", str, pkField.GetName())
	}

	str = fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n%s\n)\n", s.GetTableName(), str)
	//log.Print(str)

	ret = str

	return
}

// BuildCreateRelationTable Build CreateRelation Schema
func (s *Builder) BuildCreateRelationTable(field model.Field, rModel model.Model) (string, error) {
	relationTableName := s.GetRelationTableName(field, rModel)
	str := "\t`id` INT NOT NULL AUTO_INCREMENT,\n\t`left` INT NOT NULL,\n\t`right` INT NOT NULL,\n\tPRIMARY KEY (`id`),\n\tINDEX(`left`)"
	str = fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n%s\n)\n", relationTableName, str)
	//log.Print(str)

	return str, nil
}
