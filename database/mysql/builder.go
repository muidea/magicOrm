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

func (s *Builder) buildFilter(vModel model.Model, filter model.Filter, resultStackPtr *ResultStack) (ret string, err *cd.Result) {
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
			basicSQL, basicErr := s.buildBasicItem(field, filterItem, resultStackPtr)
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

		relationSQL, relationErr := s.buildRelationItem(vModel, field, filterItem, resultStackPtr)
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

func (s *Builder) buildBasicItem(vField model.Field, filterItem model.FilterItem, resultStackPtr *ResultStack) (ret string, err *cd.Result) {
	oprValue := filterItem.OprValue()
	oprFunc := getOprFunc(filterItem)
	oprStr, oprErr := s.buildCodec.BuildOprValue(vField, oprValue)
	if oprErr != nil {
		err = oprErr
		log.Errorf("buildBasicItem %s failed, EncodeValue error:%s", vField.GetName(), err.Error())
		return
	}

	ret = oprFunc(vField.GetName(), oprStr, resultStackPtr)
	return
}

func (s *Builder) buildRelationItem(vModel model.Model, vField model.Field, filterItem model.FilterItem, resultStackPtr *ResultStack) (ret string, err *cd.Result) {
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
	strVal := oprFunc("right", oprStr, resultStackPtr)
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

	err = cd.NewResult(cd.UnExpected, fmt.Sprintf("illegal sort field name:%s", filter.Name()))
	log.Errorf("buildSorter failed, err:%s", err.Error())
	return
}

func (s *Builder) buildFiledFilter(vField model.Field, resultStackPtr *ResultStack) (ret string, err *cd.Result) {
	fieldVal, fieldErr := s.buildCodec.BuildFieldValue(vField)
	if fieldErr != nil {
		err = fieldErr
		log.Errorf("BuildModelFilter failed, s.EncodeValue error:%s", err.Error())
		return
	}

	fieldName := vField.GetName()
	resultStackPtr.PushArgs(fieldVal.Value())
	ret = fmt.Sprintf("`%s` = ?", fieldName)
	return
}
