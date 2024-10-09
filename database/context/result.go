package context

type BuildResult interface {
	SQL() string
	Args() []any
}
