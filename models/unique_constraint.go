package models

import (
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
	known := map[string]struct{}{}
	for _, field := range fields {
		if field != nil {
			known[field.GetName()] = struct{}{}
		}
	}
	seen := map[string]struct{}{}
	for _, name := range s.Fields {
		if !constraintIdentifierPattern.MatchString(name) {
			return cd.NewError(cd.IllegalParam, "unique constraint field is invalid")
		}
		if _, ok := known[name]; !ok {
			return cd.NewError(cd.IllegalParam, "unique constraint references unavailable field")
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
type UniqueConstrainedModel interface {
	Model
	GetUniqueConstraints() []UniqueConstraint
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
