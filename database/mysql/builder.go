package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database/context"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider"
)

// Builder Builder
type Builder struct {
	common *context.Context
}

// New create builder
func New(vModel model.Model, modelProvider provider.Provider, prefix string) *Builder {
	return &Builder{
		common: context.New(vModel, modelProvider, prefix),
	}
}

func (s *Builder) buildFilter(filter model.Filter) (ret string, err *cd.Result) {
	if filter == nil {
		return
	}

	filterSQL := ""
	pkField := s.common.GetHostModelPrimaryKeyField()
	for _, field := range s.common.GetHostModelFields() {
		filterItem := filter.GetFilterItem(field.GetName())
		if filterItem == nil {
			continue
		}

		fType := field.GetType()
		if fType.IsBasic() {
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

		relationSQL, relationErr := s.buildRelationItem(pkField, field, filterItem)
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
	fType := vField.GetType()
	oprValue := filterItem.OprValue()
	oprFunc := getOprFunc(filterItem)
	oprStr, oprErr := s.common.BuildOprValue(fType, oprValue)
	if oprErr != nil {
		err = oprErr
		log.Errorf("buildBasicItem %s failed, EncodeValue error:%s", vField.GetName(), err.Error())
		return
	}

	if fType.IsBasic() {
		ret = oprFunc(vField.GetName(), oprStr)
		return
	}

	err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal item type, name:%s", vField.GetName()))
	log.Errorf("buildBasicItem failed, error:%s", err.Error())
	return
}

func (s *Builder) buildRelationItem(pkField model.Field, vField model.Field, filterItem model.FilterItem) (ret string, err *cd.Result) {
	vType := vField.GetType()
	oprValue := filterItem.OprValue()
	oprFunc := getOprFunc(filterItem)
	oprStr, oprErr := s.common.BuildOprValue(vType, oprValue)
	if oprErr != nil {
		err = oprErr
		log.Errorf("buildRelationItem %s failed, s.common.BuildOprValue error:%s", vField.GetName(), err.Error())
		return
	}

	relationFilterSQL := ""
	strVal := oprFunc("right", oprStr)
	relationTableName, relationErr := s.common.BuildRelationTableName(vField, nil)
	if relationErr != nil {
		err = relationErr
		log.Errorf("buildRelationItem %s failed, s.common.BuildRelationTableName error:%s", vField.GetName(), err.Error())
		return
	}
	relationFilterSQL = fmt.Sprintf("SELECT DISTINCT(`left`) `id`  FROM `%s` WHERE %s", relationTableName, strVal)
	relationFilterSQL = fmt.Sprintf("`%s` IN (%s)", pkField.GetName(), relationFilterSQL)
	ret = relationFilterSQL
	return
}

func (s *Builder) buildSorter(filter model.Sorter) (ret string, err *cd.Result) {
	if filter == nil {
		return
	}

	for _, field := range s.common.GetHostModelFields() {
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
	pkfVal, pkfErr := s.common.BuildFieldValue(vField.GetType(), vField.GetValue())
	if pkfErr != nil {
		err = pkfErr
		log.Errorf("BuildModelFilter failed, s.EncodeValue error:%s", err.Error())
		return
	}

	pkfName := vField.GetName()
	ret = fmt.Sprintf("`%s` = %v", pkfName, pkfVal)
	return
}
