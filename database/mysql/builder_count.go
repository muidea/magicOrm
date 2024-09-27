package mysql

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
)

// BuildCount build count
func (s *Builder) BuildCount(filter model.Filter) (ret string, err *cd.Result) {
	pkFieldName := s.common.GetHostPrimaryKeyField().GetName()
	str := fmt.Sprintf("SELECT COUNT(`%s`) FROM `%s`", pkFieldName, s.common.GetHostTableName())
	if filter != nil {
		filterSQL, filterErr := s.buildFilter(filter)
		if filterErr != nil {
			err = filterErr
			log.Errorf("BuildCount failed, s.buildFilter error:%s", err.Error())
			return
		}

		if filterSQL != "" {
			str = fmt.Sprintf("%s WHERE %s", str, filterSQL)
		}
	}

	if traceSQL() {
		log.Infof("[SQL] count: %s", str)
	}

	ret = str
	return
}
