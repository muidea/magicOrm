package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
)

// BuildQuery build query sql
func (s *Builder) BuildQuery(filter model.Filter) (ret string, err *cd.Result) {
	namesVal, nameErr := s.getFieldQueryNames(filter)
	if nameErr != nil {
		err = nameErr
		log.Errorf("BuildQuery failed, s.getFieldQueryNames error:%s", err.Error())
		return
	}

	str := fmt.Sprintf("SELECT %s FROM `%s`", namesVal, s.common.GetHostModelTableName())
	if filter != nil {
		filterSQL, filterErr := s.buildFilter(filter)
		if filterErr != nil {
			err = filterErr
			log.Errorf("BuildQuery failed, s.buildFilter error:%s", err.Error())
			return
		}

		if filterSQL != "" {
			str = fmt.Sprintf("%s WHERE %s", str, filterSQL)
		}

		sortVal, sortErr := s.buildSorter(filter.Sorter())
		if sortErr != nil {
			err = sortErr
			log.Errorf("BuildQuery failed, s.buildSorter error:%s", err.Error())
			return
		}

		if sortVal != "" {
			str = fmt.Sprintf("%s ORDER BY %s", str, sortVal)
		}

		limit, offset, paging := filter.Pagination()
		if paging {
			str = fmt.Sprintf("%s LIMIT %d OFFSET %d", str, limit, offset)
		}
	}
	if traceSQL() {
		log.Infof("[SQL] query: %s", str)
	}

	ret = str
	//log.Print(ret)
	return
}

// BuildQueryRelation build query relation sql
func (s *Builder) BuildQueryRelation(vField model.Field, rModel model.Model) (ret string, err *cd.Result) {
	leftVal, leftErr := s.common.GetHostModelValue()
	if leftErr != nil {
		err = leftErr
		log.Errorf("BuildQueryRelation failed, s.GetHostModelValue error:%s", err.Error())
		return
	}

	relationTableName := s.common.GetRelationTableName(vField, rModel)
	str := fmt.Sprintf("SELECT `right` FROM `%s` WHERE `left`= %v", relationTableName, leftVal)
	//log.Print(ret)
	if traceSQL() {
		log.Infof("[SQL] query relation: %s", str)
	}

	ret = str
	return
}

func (s *Builder) getFieldQueryNames(filter model.Filter) (ret string, err *cd.Result) {
	str := ""
	vModel := filter.MaskModel()
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

func (s *Builder) GetFieldScanDest(vField model.Field) (ret interface{}, err *cd.Result) {
	return getFieldScanDestPtr(vField)
}
