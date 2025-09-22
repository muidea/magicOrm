package builder

type Result interface {
	SQL() string
	Args() []any
}
