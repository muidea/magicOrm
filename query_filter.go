package orm

import (
	"muidea.com/magicOrm/model"
)

// queryFilter queryFilter
type queryFilter struct {
	params         map[string]interface{}
	modelInfoCache model.StructInfoCache
}

func (s *queryFilter) Add(key string, val interface{}) {
	s.params[key] = val
}

func (s *queryFilter) Builder() (ret string, err error) {
	return "", nil
}
