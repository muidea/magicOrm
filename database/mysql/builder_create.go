package mysql

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
)

// BuildCreateSchema  BuildCreateSchema
func (s *Builder) BuildCreateSchema() (ret string, err error) {
	str := ""
	for _, val := range s.modelInfo.GetFields() {
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

	if s.modelInfo.GetPrimaryField() != nil {
		fTag := s.modelInfo.GetPrimaryField().GetTag()
		str = fmt.Sprintf("%s,\n\tPRIMARY KEY (`%s`)", str, fTag.GetName())
	}

	str = fmt.Sprintf("CREATE TABLE `%s` (\n%s\n)\n", s.getHostTableName(s.modelInfo), str)
	//log.Print(str)

	ret = str

	return
}

// BuildCreateRelationSchema Build CreateRelation Schema
func (s *Builder) BuildCreateRelationSchema(fieldName string, relationInfo model.Model) (string, error) {
	str := "\t`id` INT NOT NULL AUTO_INCREMENT,\n\t`left` INT NOT NULL,\n\t`right` INT NOT NULL,\n\tPRIMARY KEY (`id`),\n\tINDEX(`left`)"
	str = fmt.Sprintf("CREATE TABLE `%s` (\n%s\n)\n", s.GetRelationTableName(fieldName, relationInfo), str)
	//log.Print(str)

	return str, nil
}
