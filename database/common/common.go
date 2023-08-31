package common

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

type Common struct {
	entityModel   model.Model
	modelProvider provider.Provider
	specialPrefix string

	// temp value, for performance optimization
	entityTableName string
	entityValue     interface{}
}

func New(vModel model.Model, modelProvider provider.Provider, prefix string) Common {
	return Common{entityModel: vModel, modelProvider: modelProvider, specialPrefix: prefix}
}

func (s *Common) constructTableName(vModel model.Model) string {
	//items := strings.Split(vModel.GetName(), ".")
	//return items[len(items)-1]
	return cases.Title(language.English).String(vModel.GetName())
}

func (s *Common) constructInfix(vFiled model.Field) string {
	return cases.Title(language.English).String(vFiled.GetName())
	//return strings.Title(vFiled.GetName())
}

func (s *Common) GetTableName() string {
	if s.entityTableName == "" {
		s.entityTableName = s.GetHostTableName(s.entityModel)
	}

	return s.entityTableName
}

func (s *Common) GetHostTableName(vModel model.Model) string {
	//tableName := s.constructTableName(vModel)
	//return fmt.Sprintf("%s_%s", s.modelProvider.Owner(), tableName)
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

	//return fmt.Sprintf("%s_%s%s2%s", s.modelProvider.Owner(), leftName, fieldName, rightName)
	tableName := fmt.Sprintf("%s%s%s%s", leftName, infixVal, getFieldRelation(vField), rightName)
	if s.specialPrefix != "" {
		tableName = fmt.Sprintf("%s_%s", s.specialPrefix, tableName)
	}

	return tableName
}

func (s *Common) GetPrimaryKeyField() model.Field {
	return s.entityModel.GetPrimaryField()
}

func (s *Common) GetFields() model.Fields {
	return s.entityModel.GetFields()
}

func (s *Common) GetModelValue() (ret interface{}, err error) {
	if s.entityValue == nil {
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

func (s *Common) GetRelationValue(rModel model.Model) (leftVal, rightVal interface{}, err error) {
	entityVal, entityErr := s.GetModelValue()
	if entityErr != nil {
		err = entityErr
		log.Errorf("get entity model value failed, err:%s", err.Error())
		return
	}

	relationVal, relationErr := s.EncodeModelValue(rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("get relation model value failed, err:%s", err.Error())
		return
	}

	leftVal = entityVal
	rightVal = relationVal
	return
}

func (s *Common) GetFieldInitializeValue(vField model.Field) (ret interface{}, err error) {
	return getFieldInitializeValue(vField)
}

func (s *Common) EncodeValue(vValue model.Value, vType model.Type) (ret interface{}, err error) {
	fStr, fErr := s.modelProvider.EncodeValue(vValue, vType)
	if fErr != nil {
		err = fErr
		return
	}

	switch vType.GetValue() {
	case model.TypeStringValue, model.TypeDateTimeValue, model.TypeSliceValue:
		ret = fmt.Sprintf("'%v'", strings.ReplaceAll(fmt.Sprintf("%v", fStr), "'", "''"))
	default:
		ret = fStr
	}

	return
}

func (s *Common) EncodeModelValue(vModel model.Model) (ret interface{}, err error) {
	pkField := vModel.GetPrimaryField()
	fStr, fErr := s.EncodeValue(pkField.GetValue(), pkField.GetType())
	if fErr != nil {
		err = fErr
		log.Errorf("encode pkField value failed, pkField name:%s, err:%s", pkField.GetName(), err.Error())
		return
	}

	ret = fStr
	return
}

func (s *Common) GetTypeModel(vType model.Type) (model.Model, error) {
	return s.modelProvider.GetTypeModel(vType)
}
