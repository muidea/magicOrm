package codec

type BuildResult interface {
	SQL() string
	Args() []any
}
