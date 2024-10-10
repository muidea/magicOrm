package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/model"
)

func (s *Builder) BuildCreateTable() (ret codec.BuildResult, err *cd.Result) {
	createSQL := ""
	for _, field := range s.hostModel.GetFields() {
		fType := field.GetType()
		if !fType.IsBasic() {
			continue
		}

		infoVal, infoErr := s.declareFieldInfo(field)
		if infoErr != nil {
			err = infoErr
			log.Errorf("BuildCreateTable failed, declareFieldInfo error:%s", err.Error())
			return
		}

		if createSQL == "" {
			createSQL = fmt.Sprintf("\t%s", infoVal)
		} else {
			createSQL = fmt.Sprintf("%s,\n\t%s", createSQL, infoVal)
		}
	}

	pkFieldName := s.hostModel.GetPrimaryField().GetName()
	createSQL = fmt.Sprintf("%s,\n\tPRIMARY KEY (`%s`)", createSQL, pkFieldName)

	createSQL = fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n%s\n)\n", s.buildContext.BuildHostModelTableName(), createSQL)
	if traceSQL() {
		log.Infof("[SQL] create: %s", createSQL)
	}

	ret = NewBuildResult(createSQL, nil)
	return
}

// BuildCreateRelationTable Build CreateRelation Schema
func (s *Builder) BuildCreateRelationTable(vField model.Field, rModel model.Model) (ret codec.BuildResult, err *cd.Result) {
	lPKField := s.hostModel.GetPrimaryField()
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

	relationTableName, relationErr := s.buildContext.BuildRelationTableName(vField, rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("BuildCreateRelationTable %s failed, s.buildContext.BuildRelationTableName error:%s", vField.GetName(), err.Error())
		return
	}

	createRelationSQL := fmt.Sprintf("\t`id` BIGINT NOT NULL AUTO_INCREMENT,\n\t`left` %s NOT NULL,\n\t`right` %s NOT NULL,\n\tPRIMARY KEY (`id`),\n\tINDEX(`left`)", lPKType, rPKType)
	createRelationSQL = fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n%s\n)\n", relationTableName, createRelationSQL)
	//log.Print(createRelationSQL)
	if traceSQL() {
		log.Infof("[SQL] create relation: %s", createRelationSQL)
	}

	ret = NewBuildResult(createRelationSQL, nil)
	return
}

func (s *Builder) declareFieldInfo(vField model.Field) (ret string, err *cd.Result) {
	autoIncrement := ""
	fSpec := vField.GetSpec()
	if fSpec != nil && model.IsAutoIncrement(fSpec.GetValueDeclare()) {
		autoIncrement = "AUTO_INCREMENT"
	}

	allowNull := "NOT NULL"
	typeVal, typeErr := getTypeDeclare(vField.GetType(), vField.GetSpec())
	if typeErr != nil {
		err = typeErr
		log.Errorf("declareFieldInfo failed, getTypeDeclare error:%s", err.Error())
		return
	}

	ret = fmt.Sprintf("`%s` %s %s %s", vField.GetName(), typeVal, allowNull, autoIncrement)
	return
}
