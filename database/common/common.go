package common

import (
	"fmt"
	"strings"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/util"
)

type Common struct {
	entityModel   model.Model
	modelProvider provider.Provider

	// temp value, for performance optimization
	entityTableName string
	entityValue     interface{}
}

func New(vModel model.Model, modelProvider provider.Provider) Common {
	return Common{entityModel: vModel, modelProvider: modelProvider}
}

func (s *Common) constructTableName(vModel model.Model) string {
	//items := strings.Split(vModel.GetName(), ".")
	//return items[len(items)-1]
	return vModel.GetName()
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
	return s.constructTableName(vModel)
}

func (s *Common) GetRelationTableName(vField model.Field, rModel model.Model) string {
	leftName := s.constructTableName(s.entityModel)
	rightName := s.constructTableName(rModel)

	//return fmt.Sprintf("%s_%s%s2%s", s.modelProvider.Owner(), leftName, fieldName, rightName)
	return fmt.Sprintf("%s%s%s%s", leftName, vField.GetName(), getFieldRelation(vField), rightName)
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

func (s *Common) GetInitializeValue(vField model.Field) (ret interface{}, err error) {
	return getFieldInitializeValue(vField)
}

func (s *Common) EncodeValue(vValue model.Value, vType model.Type) (ret interface{}, err error) {
	fStr, fErr := s.modelProvider.EncodeValue(vValue, vType)
	if fErr != nil {
		err = fErr
		return
	}

	switch vType.GetValue() {
	case util.TypeStringField, util.TypeDateTimeField, util.TypeSliceField:
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