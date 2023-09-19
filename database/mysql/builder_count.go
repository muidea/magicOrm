package mysql

import (
	"fmt"

	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
)

// BuildCount build count
func (s *Builder) BuildCount(filter model.Filter) (ret string, err error) {
	pkField := s.GetPrimaryKeyField(nil)
	str := fmt.Sprintf("SELECT COUNT(%s) FROM `%s`", pkField.GetName(), s.GetTableName())
	if filter != nil {
		filterSQL, filterErr := s.buildFilter(filter)
		if filterErr != nil {
			err = filterErr
			log.Errorf("buildModelFilter failed, err:%s", err.Error())
			return
		}

		if filterSQL != "" {
			str = fmt.Sprintf("%s WHERE %s", str, filterSQL)
		}
	}

	//log.Print(ret)

	ret = str
	return
}
