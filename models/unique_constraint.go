package models

import (
	"reflect"
	"regexp"

	cd "github.com/muidea/magicCommon/def"
)

var constraintIdentifierPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// UniqueConstraint declares a database-enforced unique key derived from a
// model definition. It is schema metadata, never a runtime SQL fragment.
type UniqueConstraint struct {
	Name   string   `json:"name"`
	Fields []string `json:"fields"`
}

// Verify checks that the constraint name and every referenced field are part
// of the active model. This keeps schema builders from interpolating arbitrary
// identifiers supplied outside a model declaration.
func (s UniqueConstraint) Verify(fields Fields) *cd.Error {
	if !constraintIdentifierPattern.MatchString(s.Name) {
		return cd.NewError(cd.IllegalParam, "unique constraint name is invalid")
	}
	if len(s.Fields) == 0 {
		return cd.NewError(cd.IllegalParam, "unique constraint fields are empty")
	}
	known := map[string]bool{}
	for _, field := range fields {
		if field != nil {
			// 单值引用参与唯一声明后由主表保存目标主键；多值/包含仍拒绝。
			fieldType := field.GetType()
			typed := fieldType != nil
			if typed && reflect.ValueOf(fieldType).Kind() == reflect.Pointer {
				typed = !reflect.ValueOf(fieldType).IsNil()
			}
			known[field.GetName()] = typed && (IsBasicField(field) || IsSingleReferenceField(field))
		}
	}
	seen := map[string]struct{}{}
	for _, name := range s.Fields {
		if !constraintIdentifierPattern.MatchString(name) {
			return cd.NewError(cd.IllegalParam, "unique constraint field is invalid")
		}
		physical, ok := known[name]
		if !ok {
			return cd.NewError(cd.IllegalParam, "unique constraint references unavailable field")
		}
		if !physical {
			return cd.NewError(cd.IllegalParam, "unique constraint requires a scalar, serialized basic collection, or single reference field")
		}
		if _, duplicate := seen[name]; duplicate {
			return cd.NewError(cd.IllegalParam, "unique constraint contains duplicate field")
		}
		seen[name] = struct{}{}
	}
	return nil
}

// UniqueConstrainedModel is implemented by dynamic remote models that carry
// schema-level composite uniqueness declarations.
type UniqueConstraintDeclarer interface {
	GetUniqueConstraints() []UniqueConstraint
}

type UniqueConstrainedModel interface {
	Model
	UniqueConstraintDeclarer
}

// GetUniqueConstraints returns a defensive copy of model-declared keys.
func GetUniqueConstraints(model Model) []UniqueConstraint {
	declared, ok := model.(UniqueConstrainedModel)
	if !ok {
		return nil
	}
	constraints := declared.GetUniqueConstraints()
	if len(constraints) == 0 {
		return nil
	}
	ret := make([]UniqueConstraint, len(constraints))
	for idx, constraint := range constraints {
		ret[idx] = UniqueConstraint{Name: constraint.Name, Fields: append([]string(nil), constraint.Fields...)}
	}
	return ret
}
