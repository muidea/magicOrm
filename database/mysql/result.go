package mysql

type Result struct {
	sqlVal  string
	valsVal []any
}

func (s *Result) String() string {
	return s.sqlVal
}

func (s *Result) SQL() string {
	return s.sqlVal
}

func (s *Result) Args() []any {
	return s.valsVal
}

func NewResult(sql string, vals []any) *Result {
	return &Result{
		sqlVal:  sql,
		valsVal: []any{},
	}
}
