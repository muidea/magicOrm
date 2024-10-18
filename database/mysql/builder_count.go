package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
)

// BuildCount build count
func (s *Builder) BuildCount(vModel model.Model, filter model.Filter) (ret *Result, err *cd.Result) {
	pkFieldName := vModel.GetPrimaryField().GetName()
	countSQL := fmt.Sprintf("SELECT COUNT(`%s`) FROM `%s`", pkFieldName, s.buildCodec.ConstructModelTableName(vModel))
	if filter != nil {
		filterSQL, filterErr := s.buildFilter(vModel, filter)
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

	ret = NewResult(countSQL, nil)
	return
}
