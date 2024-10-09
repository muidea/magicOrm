package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/database/context"
	"github.com/muidea/magicOrm/model"
)

// BuildCount build count
func (s *Builder) BuildCount(filter model.Filter) (ret context.BuildResult, err *cd.Result) {
	pkFieldName := s.hostModel.GetPrimaryField().GetName()
	countSQL := fmt.Sprintf("SELECT COUNT(`%s`) FROM `%s`", pkFieldName, s.common.BuildHostModelTableName())
	if filter != nil {
		filterSQL, filterErr := s.buildFilter(filter)
		if filterErr != nil {
			err = filterErr
			log.Errorf("BuildCount failed, s.buildFilter error:%s", err.Error())
			return
		}

		if filterSQL != "" {
			countSQL = fmt.Sprintf("%s WHERE %s", countSQL, filterSQL)
		}
	}

	if traceSQL() {
		log.Infof("[SQL] count: %s", countSQL)
	}

	ret = NewBuildResult(countSQL, nil)
	return
}
