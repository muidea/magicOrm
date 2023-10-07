package common

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

type Common struct {
	entityModel   model.Model
	modelProvider provider.Provider
	specialPrefix string

	// temp value, for performance optimization
	entityTableName string
	entityValue     string
}

func New(vModel model.Model, modelProvider provider.Provider, prefix string) Common {
	return Common{entityModel: vModel, modelProvider: modelProvider, specialPrefix: prefix}
}

func (s *Common) constructTableName(vModel model.Model) string {
	return cases.Title(language.English).String(vModel.GetName())
}

func (s *Common) constructInfix(vFiled model.Field) string {
	return cases.Title(language.English).String(vFiled.GetName())
}

func (s *Common) GetTableName() string {
	if s.entityTableName == "" {
		s.entityTableName = s.GetHostTableName(s.entityModel)
	}

	return s.entityTableName
}

func (s *Common) GetHostTableName(vModel model.Model) string {
	tableName := s.constructTableName(vModel)
	if s.specialPrefix != "" {
		tableName = fmt.Sprintf("%s_%s", s.specialPrefix, tableName)
	}

	return tableName
}

func (s *Common) GetRelationTableName(vField model.Field, rModel model.Model) string {
	leftName := s.constructTableName(s.entityModel)
	rightName := s.constructTableName(rModel)
	infixVal := s.constructInfix(vField)

	tableName := fmt.Sprintf("%s%s%s%s", leftName, infixVal, getFieldRelation(vField), rightName)
	if s.specialPrefix != "" {
		tableName = fmt.Sprintf("%s_%s", s.specialPrefix, tableName)
	}

	return tableName
}

func (s *Common) GetPrimaryKeyField(vModel model.Model) model.Field {
	if vModel == nil {
		return s.entityModel.GetPrimaryField()
	}

	return vModel.GetPrimaryField()
}

func (s *Common) GetFields() model.Fields {
	return s.entityModel.GetFields()
}

func (s *Common) GetModelValue() (ret string, err error) {
	if s.entityValue == "" {
		entityVal, entityErr := s.EncodeModelValue(s.entityModel)
		if entityErr != nil {
			err = entityErr
			return
		}

		s.entityValue = entityVal
	}

	ret = s.entityValue
	return
}

func (s *Common) GetRelationValue(rModel model.Model) (leftVal, rightVal string, err error) {
	entityVal, entityErr := s.GetModelValue()
	if entityErr != nil {
		err = entityErr
		log.Errorf("GetRelationValue failed, s.GetModelValue error:%s", err.Error())
		return
	}

	relationVal, relationErr := s.EncodeModelValue(rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("GetRelationValue failed, s.EncodeModelValue error:%s", err.Error())
		return
	}

	leftVal = entityVal
	rightVal = relationVal
	return
}

func (s *Common) EncodeValue(vValue model.Value, vType model.Type) (ret string, err error) {
	defer func() {
		if eErr := recover(); eErr != nil {
			err = fmt.Errorf("encode value failed, type:%v, err:%v", vType.GetPkgKey(), eErr)
			log.Errorf("EncodeValue failed, error:%s", err.Error())
		}
	}()

	fStr, fErr := s.modelProvider.EncodeValue(vValue, vType)
	if fErr != nil {
		err = fErr
		log.Errorf("EncodeValue failed, s.modelProvider.EncodeValue error:%s", err.Error())
		return
	}

	switch vType.GetValue() {
	case model.TypeStringValue, model.TypeDateTimeValue, model.TypeSliceValue:
		ret = fmt.Sprintf("'%v'", strings.ReplaceAll(fStr.(string), "'", "''"))
	default:
		ret = fmt.Sprintf("%v", fStr)
	}

	return
}

func (s *Common) EncodeModelValue(vModel model.Model) (ret string, err error) {
	pkField := vModel.GetPrimaryField()
	fStr, fErr := s.EncodeValue(pkField.GetValue(), pkField.GetType())
	if fErr != nil {
		err = fErr
		log.Errorf("EncodeModelValue failed, s.EncodeValue error:%s", err.Error())
		return
	}

	ret = fStr
	return
}

func (s *Common) GetTypeModel(vType model.Type) (ret model.Model, err error) {
	vModel, vErr := s.modelProvider.GetTypeModel(vType)
	if vErr != nil {
		err = vErr
		log.Errorf("GetTypeModel failed, s.modelProvider.GetTypeModel error:%s", err.Error())
		return
	}

	ret = vModel
	return
}
