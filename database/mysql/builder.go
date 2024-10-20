package mysql

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

func (s *Builder) buildFilter(vModel model.Model, filter model.Filter) (ret string, err *cd.Result) {
	if filter == nil {
		return
	}

	filterSQL := ""
	for _, field := range vModel.GetFields() {
		filterItem := filter.GetFilterItem(field.GetName())
		if filterItem == nil {
			continue
		}

		if field.IsBasic() {
			basicSQL, basicErr := s.buildBasicItem(field, filterItem)
			if basicErr != nil {
				err = basicErr
				log.Errorf("buildFilter failed, s.buildBasicItem %s error:%s", field.GetName(), err.Error())
				return
			}

			if filterSQL == "" {
				filterSQL = fmt.Sprintf("%s", basicSQL)
				continue
			}

			filterSQL = fmt.Sprintf("%s AND %s", filterSQL, basicSQL)
			continue
		}

		relationSQL, relationErr := s.buildRelationItem(vModel, field, filterItem)
		if relationErr != nil {
			err = relationErr
			log.Errorf("buildFilter failed, s.buildRelationItem %s error:%s", field.GetName(), err.Error())
			return
		}

		if filterSQL == "" {
			filterSQL = fmt.Sprintf("%s", relationSQL)
			continue
		}

		filterSQL = fmt.Sprintf("%s AND %s", filterSQL, relationSQL)
	}

	ret = filterSQL
	return
}

func (s *Builder) buildBasicItem(vField model.Field, filterItem model.FilterItem) (ret string, err *cd.Result) {
	oprValue := filterItem.OprValue()
	oprFunc := getOprFunc(filterItem)
	oprStr, oprErr := s.buildCodec.BuildOprValue(vField, oprValue)
	if oprErr != nil {
		err = oprErr
		log.Errorf("buildBasicItem %s failed, EncodeValue error:%s", vField.GetName(), err.Error())
		return
	}

	ret = oprFunc(vField.GetName(), oprStr)
	return
}

func (s *Builder) buildRelationItem(vModel model.Model, vField model.Field, filterItem model.FilterItem) (ret string, err *cd.Result) {
	oprValue := filterItem.OprValue()
	oprFunc := getOprFunc(filterItem)
	oprStr, oprErr := s.buildCodec.BuildOprValue(vField, oprValue)
	if oprErr != nil {
		err = oprErr
		log.Errorf("buildRelationItem %s failed, s.buildCodec.BuildOprValue error:%s", vField.GetName(), err.Error())
		return
	}

	rModel, rErr := s.modelProvider.GetTypeModel(vField.GetType())
	if rErr != nil {
		err = rErr
		log.Errorf("buildRelationItem %s failed, s.modelProvider.GetTypeModel error:%s", vField.GetName(), err.Error())
		return
	}

	relationFilterSQL := ""
	strVal := oprFunc("right", oprStr)
	relationTableName, relationErr := s.buildCodec.ConstructRelationTableName(vModel, vField, rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("buildRelationItem %s failed, s.buildCodec.ConstructRelationTableName error:%s", vField.GetName(), err.Error())
		return
	}

	pkField := vModel.GetPrimaryField()
	relationFilterSQL = fmt.Sprintf("SELECT DISTINCT(`left`) `id`  FROM `%s` WHERE %s", relationTableName, strVal)
	relationFilterSQL = fmt.Sprintf("`%s` IN (%s)", pkField.GetName(), relationFilterSQL)
	ret = relationFilterSQL
	return
}

func (s *Builder) buildSorter(vModel model.Model, filter model.Sorter) (ret string, err *cd.Result) {
	if filter == nil {
		return
	}

	for _, field := range vModel.GetFields() {
		if field.GetName() == filter.Name() {
			ret = SortOpr(filter.Name(), filter.AscSort())
			return
		}
	}

	err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal sort field name:%s", filter.Name()))
	log.Errorf("buildSorter failed, err:%s", err.Error())
	return
}

func (s *Builder) buildFiledFilter(vField model.Field) (ret string, err *cd.Result) {
	fieldVal, fieldErr := s.buildCodec.BuildFieldValue(vField)
	if fieldErr != nil {
		err = fieldErr
		log.Errorf("BuildModelFilter failed, s.EncodeValue error:%s", err.Error())
		return
	}

	fieldName := vField.GetName()
	ret = fmt.Sprintf("`%s` = %v", fieldName, fieldVal)
	return
}
