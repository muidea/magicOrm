package mysql

import (
	"fmt"
	"strings"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/util"
)

// Builder Builder
type Builder struct {
	entityModel   model.Model
	modelProvider provider.Provider

	// temp value, for performance optimization
	entityTableName string
	entityValue     interface{}
}

// New create builder
func New(vModel model.Model, modelProvider provider.Provider) *Builder {
	return &Builder{entityModel: vModel, modelProvider: modelProvider}
}

func (s *Builder) constructTableName(vModel model.Model) string {
	//items := strings.Split(vModel.GetName(), ".")
	//return items[len(items)-1]
	return vModel.GetName()
}

// GetTableName GetTableName
func (s *Builder) GetTableName() string {
	if s.entityTableName == "" {
		s.entityTableName = s.getHostTableName(s.entityModel)
	}

	return s.entityTableName
}

// getHostTableName getHostTableName
func (s *Builder) getHostTableName(vModel model.Model) string {
	//tableName := s.constructTableName(vModel)
	//return fmt.Sprintf("%s_%s", s.modelProvider.Owner(), tableName)
	return s.constructTableName(vModel)
}

func (s *Builder) GetRelationTableName(vField model.Field, rModel model.Model) string {
	leftName := s.constructTableName(s.entityModel)
	rightName := s.constructTableName(rModel)

	//return fmt.Sprintf("%s_%s%s2%s", s.modelProvider.Owner(), leftName, fieldName, rightName)
	return fmt.Sprintf("%s%s%s%s", leftName, vField.GetName(), getFieldRelation(vField), rightName)
}

func (s *Builder) getModelValue() (ret interface{}, err error) {
	if s.entityValue == nil {
		entityVal, entityErr := s.encodeModelValue(s.entityModel)
		if entityErr != nil {
			err = entityErr
			return
		}
		s.entityValue = entityVal
	}

	ret = s.entityValue
	return
}

func (s *Builder) getRelationValue(rModel model.Model) (leftVal, rightVal interface{}, err error) {
	entityVal, entityErr := s.getModelValue()
	if entityErr != nil {
		err = entityErr
		log.Errorf("get entity model value failed, err:%s", err.Error())
		return
	}

	relationVal, relationErr := s.encodeModelValue(rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("get relation model value failed, err:%s", err.Error())
		return
	}

	leftVal = entityVal
	rightVal = relationVal
	return
}

func (s *Builder) GetInitializeValue(vField model.Field) (ret interface{}, err error) {
	return getFieldInitializeValue(vField)
}

func (s *Builder) encodeValue(vValue model.Value, vType model.Type) (ret interface{}, err error) {
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

func (s *Builder) encodeModelValue(vModel model.Model) (ret interface{}, err error) {
	pkField := vModel.GetPrimaryField()
	fStr, fErr := s.encodeValue(pkField.GetValue(), pkField.GetType())
	if fErr != nil {
		err = fErr
		log.Errorf("encode pkField value failed, pkField name:%s, err:%s", pkField.GetName(), err.Error())
		return
	}

	ret = fStr
	return
}

func (s *Builder) buildModelFilter(vModel model.Model) (ret string, err error) {
	pkfVal, pkfErr := s.encodeModelValue(vModel)
	if pkfErr != nil {
		err = pkfErr
		return
	}

	pkfTag := vModel.GetPrimaryField().GetTag().GetName()
	ret = fmt.Sprintf("`%s`=%v", pkfTag, pkfVal)
	return
}
