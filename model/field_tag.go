package model

// FieldTag FieldTag
type FieldTag interface {
	Name() string
	IsPrimaryKey() bool
	IsAutoIncrement() bool
	String() string
	Copy() FieldTag
}
