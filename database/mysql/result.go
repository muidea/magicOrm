package mysql

import (
	"github.com/muidea/magicOrm/database/codec"
)

type BuildResult struct {
	sqlVal  string
	valsVal []any
}

func (s *BuildResult) String() string {
	return s.sqlVal
}

func (s *BuildResult) SQL() string {
	return s.sqlVal
}

func (s *BuildResult) Args() []any {
	return s.valsVal
}

func NewBuildResult(sql string, vals []any) codec.BuildResult {
	return &BuildResult{
		sqlVal:  sql,
		valsVal: []any{},
	}
}
