package model

// Tag Tag
type Tag interface {
	GetName() string
	IsPrimaryKey() bool
	IsAutoIncrement() bool
}
