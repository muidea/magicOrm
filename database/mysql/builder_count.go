package mysql

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
)

// BuildCount BuildCount
func (s *Builder) BuildCount(filter model.Filter) (ret string, err error) {
	pkField := s.modelInfo.GetPrimaryField()

	ret = fmt.Sprintf("SELECT COUNT(%s) FROM `%s`", pkField.GetTag().GetName(), s.GetHostTableName(s.modelInfo))
	if filter != nil {
		filterSQL, filterErr := s.buildBatchFilter(filter)
		if filterErr != nil {
			err = filterErr
			return
		}

		if filterSQL != "" {
			ret = fmt.Sprintf("%s WHERE %s", ret, filterSQL)
		}
	}

	//log.Print(ret)
	return
}
