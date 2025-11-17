package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/models"
)

// BuildQuery build query sql
func (s *Builder) BuildQuery(vModel models.Model, filter models.Filter) (ret database.Result, err *cd.Error) {
	namesVal, nameErr := s.getFieldQueryNames(vModel)
	if nameErr != nil {
		err = nameErr
		log.Errorf("BuildQuery failed, s.getFieldQueryNames error:%s", err.Error())
		return
	}

	resultStackPtr := &ResultStack{}
	querySQL := fmt.Sprintf("SELECT %s FROM `%s`", namesVal, s.buildCodec.ConstructModelTableName(vModel))
	if filter != nil {
		filterSQL, filterErr := s.buildFilter(vModel, filter, resultStackPtr)
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

		paginationer := filter.Paginationer()
		if paginationer != nil {
			resultStackPtr.PushArgs(paginationer.Limit(), paginationer.Offset())
			querySQL = fmt.Sprintf("%s LIMIT ? OFFSET ?", querySQL)
		}
	}
	if traceSQL() {
		log.Infof("[SQL] query: %s", querySQL)
	}

	resultStackPtr.SetSQL(querySQL)
	ret = resultStackPtr
	return
}

// BuildQueryRelation build query relation sql
func (s *Builder) BuildQueryRelation(vModel models.Model, vField models.Field) (ret database.Result, err *cd.Error) {
	leftVal := vModel.GetPrimaryField().GetValue().Get()
	relationTableName, relationErr := s.buildCodec.ConstructRelationTableName(vModel, vField)
	if relationErr != nil {
		err = relationErr
		log.Errorf("BuildQueryRelation %s failed, s.buildCodec.ConstructRelationTableName error:%s", vField.GetName(), err.Error())
		return
	}

	resultStackPtr := &ResultStack{}
	queryRelationSQL := fmt.Sprintf("SELECT `right` FROM `%s` WHERE `left`= ?", relationTableName)
	if traceSQL() {
		log.Infof("[SQL] query relation: %s", queryRelationSQL)
	}

	resultStackPtr.PushArgs(leftVal)
	resultStackPtr.SetSQL(queryRelationSQL)
	ret = resultStackPtr
	return
}

func (s *Builder) getFieldQueryNames(vModel models.Model) (ret string, err *cd.Error) {
	str := ""
	for _, field := range vModel.GetFields() {
		if !models.IsBasicField(field) || !models.IsValidField(field) {
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

func (s *Builder) BuildModuleValueHolder(vModel models.Model) (ret []any, err *cd.Error) {
	items := []any{}
	for _, field := range vModel.GetFields() {
		if !models.IsBasicField(field) || !models.IsValidField(field) {
			continue
		}

		itemVal, itemErr := getFieldPlaceHolder(field.GetType())
		if itemErr != nil {
			err = itemErr
			log.Errorf("BuildModuleValueHolder failed, getFieldPlaceHolder error:%s", err.Error())
			return
		}

		items = append(items, itemVal)
	}

	ret = items
	return
}
