package mysql

import (
	"fmt"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/model"
)

// BuildCount build count
func (s *Builder) BuildCount(filter model.Filter) (ret string, err error) {
	pkField := s.GetPrimaryKeyField(nil)
	ret = fmt.Sprintf("SELECT COUNT(%s) FROM `%s`", pkField.GetName(), s.GetTableName())
	if filter != nil {
		filterSQL, filterErr := s.buildFilter(filter)
		if filterErr != nil {
			err = filterErr
			log.Errorf("buildModelFilter failed, err:%s", err.Error())
			return
		}

		if filterSQL != "" {
			ret = fmt.Sprintf("%s WHERE %s", ret, filterSQL)
		}
	}

	//log.Print(ret)
	return
}
