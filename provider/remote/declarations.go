package remote

import (
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/utils"
)

type Field struct {
	Name        string     `json:"name"`
	ShowName    string     `json:"showName"`
	Description string     `json:"description"`
	Type        *TypeImpl  `json:"type"`
	Spec        *SpecImpl  `json:"spec"`
	value       *ValueImpl `json:"-"`

	valueValidator models.ValueValidator `json:"-"`
}

type FieldValue struct {
	Name     string `json:"name"`
	Value    any    `json:"value"`
	Assigned bool   `json:"assigned,omitempty"`
}

type ObjectFilter struct {
	Name           string            `json:"name"`
	PkgPath        string            `json:"pkgPath"`
	EqualFilter    []*FieldValue     `json:"equal"`
	NotEqualFilter []*FieldValue     `json:"noEqual"`
	BelowFilter    []*FieldValue     `json:"below"`
	AboveFilter    []*FieldValue     `json:"above"`
	InFilter       []*FieldValue     `json:"in"`
	NotInFilter    []*FieldValue     `json:"notIn"`
	LikeFilter     []*FieldValue     `json:"like"`
	MaskValue      *ObjectValue      `json:"maskValue"`
	PageFilter     *utils.Pagination `json:"page"`
	SortFilter     *utils.SortFilter `json:"sort"`

	bindObject *Object `json:"-"`
}

type Object struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	ShowName    string   `json:"showName"`
	Icon        string   `json:"icon"`
	PkgPath     string   `json:"pkgPath"`
	Description string   `json:"description"`
	Fields      []*Field `json:"fields"`
	// UniqueConstraints are schema declarations applied by magicOrm when the
	// model is created. They are not query filters or caller-provided SQL.
	UniqueConstraints []models.UniqueConstraint `json:"uniqueConstraints,omitempty"`
	// Indexes are non-unique lifecycle indexes for state/expiry queries.
	Indexes []models.Index `json:"indexes,omitempty"`

	// 临时变量不进行数据序列化传递
	valueValidator models.ValueValidator `json:"-"`
	viewSpec       models.ViewDeclare    `json:"-"`
	// privateReadFields is intentionally process-local. It is never serialized
	// in VMI/entity metadata and can only be enabled by a trusted storage path.
	privateReadFields map[string]struct{} `json:"-"`
}

// ObjectValue Object value
type ObjectValue struct {
	ID      string        `json:"id"`
	Name    string        `json:"name"`
	PkgPath string        `json:"pkgPath"`
	Fields  []*FieldValue `json:"fields"`
}

// SliceObjectValue slice object value
type SliceObjectValue struct {
	Name    string         `json:"name"`
	PkgPath string         `json:"pkgPath"`
	Values  []*ObjectValue `json:"values"`
}

type SpecImpl struct {
	FieldName    string               `json:"fieldName"`
	PrimaryKey   bool                 `json:"primaryKey"`
	ValueDeclare models.ValueDeclare  `json:"valueDeclare"`
	ViewDeclare  []models.ViewDeclare `json:"viewDeclare"`

	Constraint   string `json:"constraint"`
	DefaultValue any    `json:"defaultValue"`

	// 这里是为了避免在使用时多次解析
	constraintsVal models.Constraints
}

type TypeImpl struct {
	Name        string             `json:"name"`
	PkgPath     string             `json:"pkgPath"`
	Description string             `json:"description"`
	Value       models.TypeDeclare `json:"-"`
	IsPtr       bool               `json:"isPtr"`
	ElemType    *TypeImpl          `json:"elemType"`
}
