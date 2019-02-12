package mysql

import (
	"fmt"
	"log"

	"muidea.com/magicOrm/model"
)

// BuildCreateSchema  BuildCreateSchema
func (s *Builder) BuildCreateSchema() (string, error) {
	str := ""
	for _, val := range s.modelInfo.GetFields() {
		fType := val.GetType()
		dependType := fType.GetDepend()
		if dependType != nil {
			continue
		}

		if str == "" {
			str = fmt.Sprintf("\t%s", declareFieldInfo(val))
		} else {
			str = fmt.Sprintf("%s,\n\t%s", str, declareFieldInfo(val))
		}
	}
	if s.modelInfo.GetPrimaryField() != nil {
		fTag := s.modelInfo.GetPrimaryField().GetTag()
		str = fmt.Sprintf("%s,\n\tPRIMARY KEY (`%s`)", str, fTag.GetName())
	}

	str = fmt.Sprintf("CREATE TABLE `%s` (\n%s\n)\n", s.getTableName(s.modelInfo), str)
	log.Print(str)

	return str, nil
}

// BuildCreateRelationSchema BuildCreateRelationSchema
func (s *Builder) BuildCreateRelationSchema(fieldName string, relationInfo model.Model) (string, error) {
	str := "\t`id` INT NOT NULL AUTO_INCREMENT,\n\t`left` INT NOT NULL,\n\t`right` INT NOT NULL,\n\tPRIMARY KEY (`id`)"
	str = fmt.Sprintf("CREATE TABLE `%s` (\n%s\n)\n", s.GetRelationTableName(fieldName, relationInfo), str)
	log.Print(str)

	return str, nil
}
