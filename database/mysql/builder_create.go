package mysql

import (
	"fmt"
	"github.com/muidea/magicCommon/foundation/log"

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
			log.Errorf("BuildCreateTable failed, declareFieldInfo error:%s", err.Error())
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
	lPKType, lPKErr := getTypeDeclare(lPKField.GetType(), lPKField.GetSpec())
	if lPKErr != nil {
		err = lPKErr
		log.Errorf("BuildCreateRelationTable failed, getTypeDeclare error:%s", err.Error())
		return
	}

	rPKField := s.GetPrimaryKeyField(rModel)
	rPKType, rPKErr := getTypeDeclare(rPKField.GetType(), lPKField.GetSpec())
	if rPKErr != nil {
		err = rPKErr
		log.Errorf("BuildCreateRelationTable failed, getTypeDeclare error:%s", err.Error())
		return
	}

	relationTableName := s.GetRelationTableName(field, rModel)
	str := fmt.Sprintf("\t`id` BIGINT NOT NULL AUTO_INCREMENT,\n\t`left` %s NOT NULL,\n\t`right` %s NOT NULL,\n\tPRIMARY KEY (`id`),\n\tINDEX(`left`)", lPKType, rPKType)
	str = fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n%s\n)\n", relationTableName, str)
	//log.Print(str)

	ret = str
	return
}
