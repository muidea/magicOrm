package mysql

import (
	"fmt"
)

// BuildCreateSchema  BuildCreateSchema
func (s *Builder) BuildCreateSchema() (ret string, err error) {
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
		fTag := pkField.GetTag()
		str = fmt.Sprintf("%s,\n\tPRIMARY KEY (`%s`)", str, fTag.GetName())
	}

	str = fmt.Sprintf("CREATE TABLE `%s` (\n%s\n)\n", s.GetTableName(), str)
	//log.Print(str)

	ret = str

	return
}

// BuildCreateRelationSchema Build CreateRelation Schema
func (s *Builder) BuildCreateRelationSchema(relationSchema string) (string, error) {
	str := "\t`id` INT NOT NULL AUTO_INCREMENT,\n\t`left` INT NOT NULL,\n\t`right` INT NOT NULL,\n\tPRIMARY KEY (`id`),\n\tINDEX(`left`)"
	str = fmt.Sprintf("CREATE TABLE `%s` (\n%s\n)\n", relationSchema, str)
	//log.Print(str)

	return str, nil
}
