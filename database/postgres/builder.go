package postgres

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

// Builder Builder
type Builder struct {
	modelProvider provider.Provider
	buildCodec    codec.Codec
}

// New create builder
func New(provider provider.Provider, codec codec.Codec) *Builder {
	return &Builder{
		modelProvider: provider,
		buildCodec:    codec,
	}
}

func (s *Builder) buildFilter(vModel model.Model, filter model.Filter, resultStackPtr *ResultStack) (ret string, err *cd.Error) {
	if filter == nil {
		return
	}

	filterSQL := ""
	for _, field := range vModel.GetFields() {
		filterItem := filter.GetFilterItem(field.GetName())
		if filterItem == nil {
			continue
		}

		if model.IsBasicField(field) {
			basicSQL, basicErr := s.buildBasicFilterItem(field, filterItem, resultStackPtr)
			if basicErr != nil {
				err = basicErr
				log.Errorf("buildFilter failed, s.buildBasicItem %s error:%s", field.GetName(), err.Error())
				return
			}

			if filterSQL == "" {
				filterSQL = basicSQL
				continue
			}

			filterSQL = fmt.Sprintf("%s AND %s", filterSQL, basicSQL)
			continue
		}

		relationSQL, relationErr := s.buildRelationFilterItem(vModel, field, filterItem, resultStackPtr)
		if relationErr != nil {
			err = relationErr
			log.Errorf("buildFilter failed, s.buildRelationItem %s error:%s", field.GetName(), err.Error())
			return
		}

		if filterSQL == "" {
			filterSQL = relationSQL
			continue
		}

		filterSQL = fmt.Sprintf("%s AND %s", filterSQL, relationSQL)
	}

	ret = filterSQL
	return
}

func (s *Builder) buildBasicFilterItem(vField model.Field, filterItem model.FilterItem, resultStackPtr *ResultStack) (ret string, err *cd.Error) {
	oprValue := filterItem.OprValue()
	oprFunc := getOprFunc(filterItem)
	fieldVal, fieldErr := s.modelProvider.EncodeValue(oprValue.Get(), vField.GetType())
	if fieldErr != nil {
		err = fieldErr
		log.Errorf("buildBasicItem %s failed, s.modelProvider.EncodeValue error:%s", vField.GetName(), err.Error())
		return
	}

	ret = oprFunc(vField.GetName(), fieldVal, resultStackPtr)
	return
}

func (s *Builder) buildRelationFilterItem(vModel model.Model, vField model.Field, filterItem model.FilterItem, resultStackPtr *ResultStack) (ret string, err *cd.Error) {
	oprValue := filterItem.OprValue()
	oprFunc := getOprFunc(filterItem)

	var fieldVal any
	switch filterItem.OprCode() {
	case model.InOpr, model.NotInOpr:
		entitySlice := []any{}
		fieldVals := oprValue.UnpackValue()
		for _, val := range fieldVals {
			subItemVal, subItemErr := s.modelProvider.EncodeValue(val.Get(), vField.GetType())
			if subItemErr != nil {
				err = subItemErr
				log.Errorf("buildRelationItem %s failed, s.modelProvider.EncodeValue error:%s", vField.GetName(), err.Error())
				return
			}
			entitySlice = append(entitySlice, subItemVal)
		}
		fieldVal = entitySlice
	default:
		fieldVal, err = s.modelProvider.EncodeValue(oprValue.Get(), vField.GetType())
		if err != nil {
			log.Errorf("buildRelationItem %s failed, s.modelProvider.EncodeValue error:%s", vField.GetName(), err.Error())
			return
		}
	}

	relationFilterSQL := ""
	strVal := oprFunc("right", fieldVal, resultStackPtr)
	relationTableName, relationErr := s.buildCodec.ConstructRelationTableName(vModel, vField)
	if relationErr != nil {
		err = relationErr
		log.Errorf("buildRelationItem %s failed, s.buildCodec.ConstructRelationTableName error:%s", vField.GetName(), err.Error())
		return
	}

	pkField := vModel.GetPrimaryField()
	relationFilterSQL = fmt.Sprintf("SELECT DISTINCT(\"left\") \"id\"  FROM \"%s\" WHERE %s", relationTableName, strVal)
	relationFilterSQL = fmt.Sprintf("\"%s\" IN (%s)", pkField.GetName(), relationFilterSQL)
	ret = relationFilterSQL
	return
}

func (s *Builder) buildSorter(vModel model.Model, filter model.Sorter) (ret string, err *cd.Error) {
	if filter == nil {
		return
	}

	for _, field := range vModel.GetFields() {
		if field.GetName() == filter.Name() {
			ret = SortOpr(filter.Name(), filter.AscSort())
			return
		}
	}

	err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal sort field name:%s", filter.Name()))
	log.Errorf("buildSorter failed, err:%s", err.Error())
	return
}

func (s *Builder) buildFieldFilter(vField model.Field, resultStackPtr *ResultStack) (ret string, err *cd.Error) {
	fieldName := vField.GetName()
	resultStackPtr.PushArgs(vField.GetValue().Get())
	ret = fmt.Sprintf("\"%s\" = ?", fieldName)
	return
}
