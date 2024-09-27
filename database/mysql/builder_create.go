package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
)

func (s *Builder) BuildCreateTable() (ret string, err *cd.Result) {
	str := ""
	for _, val := range s.common.GetHostFields() {
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

	pkFieldName := s.common.GetHostPrimaryKeyField().GetName()
	str = fmt.Sprintf("%s,\n\tPRIMARY KEY (`%s`)", str, pkFieldName)

	str = fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n%s\n)\n", s.common.GetHostTableName(), str)
	if traceSQL() {
		log.Infof("[SQL] create: %s", str)
	}

	ret = str
	return
}

// BuildCreateRelationTable Build CreateRelation Schema
func (s *Builder) BuildCreateRelationTable(field model.Field, rModel model.Model) (ret string, err *cd.Result) {
	lPKField := s.common.GetHostPrimaryKeyField()
	lPKType, lPKErr := getTypeDeclare(lPKField.GetType(), lPKField.GetSpec())
	if lPKErr != nil {
		err = lPKErr
		log.Errorf("BuildCreateRelationTable failed, getTypeDeclare error:%s", err.Error())
		return
	}

	rPKField := rModel.GetPrimaryField()
	rPKType, rPKErr := getTypeDeclare(rPKField.GetType(), rPKField.GetSpec())
	if rPKErr != nil {
		err = rPKErr
		log.Errorf("BuildCreateRelationTable failed, getTypeDeclare error:%s", err.Error())
		return
	}

	relationTableName := s.common.GetRelationTableName(field, rModel)
	str := fmt.Sprintf("\t`id` BIGINT NOT NULL AUTO_INCREMENT,\n\t`left` %s NOT NULL,\n\t`right` %s NOT NULL,\n\tPRIMARY KEY (`id`),\n\tINDEX(`left`)", lPKType, rPKType)
	str = fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n%s\n)\n", relationTableName, str)
	//log.Print(str)
	if traceSQL() {
		log.Infof("[SQL] create relation: %s", str)
	}

	ret = str
	return
}
