package models

// IsSingleReferenceField 区分单值引用与包含对象、多值关系；不改变业务字段类型。
func IsSingleReferenceField(field Field) bool {
	return field != nil && field.GetType() != nil && IsStructField(field) && IsPtrField(field)
}

// IsInlineReference 参与声明唯一键的单值引用仅存于主表，不再维护链接表。
func IsInlineReference(model Model, field Field) bool {
	if model == nil || !IsSingleReferenceField(field) {
		return false
	}
	for _, constraint := range GetUniqueConstraints(model) {
		for _, name := range constraint.Fields {
			if name == field.GetName() {
				return true
			}
		}
	}
	return false
}

func IsMainTableField(model Model, field Field) bool {
	return IsBasicField(field) || IsInlineReference(model, field)
}
