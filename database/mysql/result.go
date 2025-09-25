package mysql

import "github.com/muidea/magicOrm/database"

type ResultStack struct {
	sqlVal  string
	argsVal []any
}

func NewError(sql string, args []any) database.Result {
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
