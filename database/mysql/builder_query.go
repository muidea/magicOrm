package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
)

// BuildQuery build query sql
func (s *Builder) BuildQuery(vModel model.Model, filter model.Filter) (ret *Result, err *cd.Result) {
	namesVal, nameErr := s.getFieldQueryNames(vModel)
	if nameErr != nil {
		err = nameErr
		log.Errorf("BuildQuery failed, s.getFieldQueryNames error:%s", err.Error())
		return
	}

	querySQL := fmt.Sprintf("SELECT %s FROM `%s`", namesVal, s.buildCodec.ConstructModelTableName(vModel))
	if filter != nil {
		filterSQL, filterErr := s.buildFilter(vModel, filter)
		if filterErr != nil {
			err = filterErr
			log.Errorf("BuildQuery failed, s.buildFilter error:%s", err.Error())
			return
		}

		if filterSQL != "" {
			querySQL = fmt.Sprintf("%s WHERE %s", querySQL, filterSQL)
		}

		sortVal, sortErr := s.buildSorter(vModel, filter.Sorter())
		if sortErr != nil {
			err = sortErr
			log.Errorf("BuildQuery failed, s.buildSorter error:%s", err.Error())
			return
		}

		if sortVal != "" {
			querySQL = fmt.Sprintf("%s ORDER BY %s", querySQL, sortVal)
		}

		limit, offset, paging := filter.Pagination()
		if paging {
			querySQL = fmt.Sprintf("%s LIMIT %d OFFSET %d", querySQL, limit, offset)
		}
	}
	if traceSQL() {
		log.Infof("[SQL] query: %s", querySQL)
	}

	ret = NewResult(querySQL, nil)
	//log.Print(ret)
	return
}

// BuildQueryRelation build query relation sql
func (s *Builder) BuildQueryRelation(vModel model.Model, vField model.Field, rModel model.Model) (ret *Result, err *cd.Result) {
	leftVal, leftErr := s.buildCodec.BuildModelValue(vModel)
	if leftErr != nil {
		err = leftErr
		log.Errorf("BuildQueryRelation failed, s.BuildHostModelValue error:%s", err.Error())
		return
	}

	relationTableName, relationErr := s.buildCodec.ConstructRelationTableName(vModel, vField, rModel)
	if relationErr != nil {
		err = relationErr
		log.Errorf("BuildQueryRelation %s failed, s.buildCodec.ConstructRelationTableName error:%s", vField.GetName(), err.Error())
		return
	}

	queryRelationSQL := fmt.Sprintf("SELECT `right` FROM `%s` WHERE `left`= %v", relationTableName, leftVal)
	//log.Print(ret)
	if traceSQL() {
		log.Infof("[SQL] query relation: %s", queryRelationSQL)
	}

	ret = NewResult(queryRelationSQL, nil)
	return
}

func (s *Builder) getFieldQueryNames(vModel model.Model) (ret string, err *cd.Result) {
	str := ""
	for _, field := range vModel.GetFields() {
		fType := field.GetType()
		fValue := field.GetValue()
		if !fType.IsBasic() || !fValue.IsValid() {
			continue
		}

		if str == "" {
			str = fmt.Sprintf("`%s`", field.GetName())
		} else {
			str = fmt.Sprintf("%s,`%s`", str, field.GetName())
		}
	}

	ret = str
	return
}

func (s *Builder) GetFieldPlaceHolder(vField model.Field) (ret any, err *cd.Result) {
	return getFieldPlaceHolder(vField)
}

func (s *Builder) BuildQueryPlaceHolder(vModel model.Model) (ret []any, err *cd.Result) {
	items := []any{}
	for _, field := range vModel.GetFields() {
		fType := field.GetType()
		fValue := field.GetValue()
		if !fType.IsBasic() || !fValue.IsValid() {
			continue
		}

		itemVal, itemErr := getFieldPlaceHolder(field)
		if itemErr != nil {
			err = itemErr
			log.Errorf("BuildQueryPlaceHolder failed, getFieldPlaceHolder error:%s", err.Error())
			return
		}

		items = append(items, itemVal)
	}

	ret = items
	return
}

func (s *Builder) BuildQueryRelationPlaceHolder(vModel model.Model, vField model.Field, rModel model.Model) (ret any, err *cd.Result) {
	return getFieldPlaceHolder(rModel.GetPrimaryField())
}
