package model

// FieldTag FieldTag
type FieldTag interface {
	GetName() string
	IsPrimaryKey() bool
	IsAutoIncrement() bool
	Copy() FieldTag
}
