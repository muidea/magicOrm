package mysql

type ResultStack struct {
	sqlVal  string
	argsVal []any
}

func NewResult(sql string, args []any) *ResultStack {
	return &ResultStack{
		sqlVal:  sql,
		argsVal: args,
	}
}

func (s *ResultStack) SetSQL(sql string) {
	s.sqlVal = sql
}

func (s *ResultStack) PushArgs(arg ...any) {
	s.argsVal = append(s.argsVal, arg...)
}

func (s *ResultStack) SQL() string {
	return s.sqlVal
}

func (s *ResultStack) Args() []any {
	return s.argsVal
}
