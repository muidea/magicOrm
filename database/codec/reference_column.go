package codec

import (
	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/utils"
)

// ReferenceColumn 将引用声明投影为目标主键列；投影不写入模型或客户端 DTO。
func ReferenceColumn(p provider.Provider, field models.Field) (models.Field, *cd.Error) {
	if !models.IsSingleReferenceField(field) {
		return nil, cd.NewError(cd.IllegalParam, "主表引用列需要单值引用")
	}
	target, err := p.GetTypeModel(field.GetType())
	if err != nil {
		return nil, err
	}
	if target == nil {
		return nil, cd.NewError(cd.IllegalParam, "引用目标模型不可用")
	}
	primary := target.GetPrimaryField()
	if primary == nil || !models.IsBasicField(primary) || models.IsSliceField(primary) {
		return nil, cd.NewError(cd.IllegalParam, "引用目标需要标量主键")
	}
	return referenceColumn{Field: field, primary: primary}, nil
}

type referenceColumn struct {
	models.Field
	primary models.Field
}

func (s referenceColumn) GetType() models.Type {
	return referenceColumnType{Type: s.primary.GetType()}
}
func (s referenceColumn) GetSpec() models.Spec {
	return referenceColumnSpec{Spec: s.Field.GetSpec()}
}

type referenceColumnType struct{ models.Type }

func (s referenceColumnType) IsPtrType() bool   { return true }
func (s referenceColumnType) Elem() models.Type { return s }

type referenceColumnSpec struct{ models.Spec }

func (s referenceColumnSpec) IsPrimaryKey() bool                   { return false }
func (s referenceColumnSpec) GetValueDeclare() models.ValueDeclare { return models.Customer }
func (s referenceColumnSpec) GetDefaultValue() any                 { return nil }

// ValidateUniqueColumns 确认补建约束引用的每个字段确实属于当前主表。
func ValidateUniqueColumns(p provider.Provider, model models.Model, constraint models.UniqueConstraint) *cd.Error {
	if err := constraint.Verify(model.GetFields()); err != nil {
		return err
	}
	for _, name := range constraint.Fields {
		field := model.GetField(name)
		if !models.IsMainTableField(model, field) {
			return cd.NewError(cd.IllegalParam, "唯一约束字段没有主表存储列")
		}
		if models.IsInlineReference(model, field) {
			if _, err := ReferenceColumn(p, field); err != nil {
				return err
			}
		}
	}
	return nil
}

// PackMainTableValue 与普通标量共用编码；引用只取已有非零主键，不能创建目标。
func PackMainTableValue(p provider.Provider, c Codec, model models.Model, field models.Field) (any, *cd.Error) {
	if !models.IsInlineReference(model, field) {
		return c.PackedBasicFieldValue(field, field.GetValue())
	}
	if utils.IsReallyNil(field.GetValue().Get()) {
		return nil, nil
	}
	target, err := p.GetTypeModel(field.GetType())
	if err != nil {
		return nil, err
	}
	target, err = p.SetModelValue(target, field.GetValue())
	if err != nil {
		return nil, err
	}
	if target == nil {
		return nil, cd.NewError(cd.IllegalParam, "引用目标模型不可用")
	}
	primary := target.GetPrimaryField()
	if primary == nil || !models.IsValidField(primary) || primary.GetValue().IsZero() {
		return nil, cd.NewError(cd.IllegalParam, "主表引用缺少目标主键")
	}
	return c.PackedBasicFieldValue(primary, primary.GetValue())
}
