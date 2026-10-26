package models

import (
	"fmt"
	cd "github.com/muidea/magicCommon/def"
)

// Index declares a non-unique database index derived from a model definition.
// It is schema metadata only; callers cannot supply arbitrary SQL fragments.
type Index struct {
	Name   string   `json:"name"`
	Fields []string `json:"fields"`
}

// Verify checks that the index name and fields are safe identifiers declared
// by the active model.
func (s Index) Verify(fields Fields) *cd.Error {
	if !constraintIdentifierPattern.MatchString(s.Name) {
		return cd.NewError(cd.IllegalParam, "index name is invalid")
	}
	if len(s.Fields) == 0 {
		return cd.NewError(cd.IllegalParam, "index fields are empty")
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
			return cd.NewError(cd.IllegalParam, "index field is invalid")
		}
		if _, ok := known[name]; !ok {
			return cd.NewError(cd.IllegalParam, "index references unavailable field")
		}
		if _, duplicate := seen[name]; duplicate {
			return cd.NewError(cd.IllegalParam, "index contains duplicate field")
		}
		seen[name] = struct{}{}
	}
	return nil
}

// IndexedModel is implemented by dynamic remote models carrying lifecycle
// indexes. Like unique constraints, indexes follow the entity model rather
// than deployment-managed SQL.
type IndexedModel interface {
	Model
	GetIndexes() []Index
}

// GetIndexes returns a defensive copy of model-declared indexes.
func GetIndexes(model Model) []Index {
	declared, ok := model.(IndexedModel)
	if !ok {
		return nil
	}
	indexes := declared.GetIndexes()
	if len(indexes) == 0 {
		return nil
	}
	ret := make([]Index, len(indexes))
	for idx, index := range indexes {
		ret[idx] = Index{Name: index.Name, Fields: append([]string(nil), index.Fields...)}
	}
	return ret
}

// VerifyModel checks the physical columns of a main-table index. Relation-table
// fields cannot be indexed on their host; declaring an index never changes storage.
// Registration/inspection retain Verify so old declarations remain diagnosable.
func (s Index) VerifyModel(model Model) *cd.Error {
	if model == nil {
		return cd.NewError(cd.IllegalParam, "index model is unavailable")
	}
	if err := s.Verify(model.GetFields()); err != nil {
		return err
	}
	for _, name := range s.Fields {
		for _, field := range model.GetFields() {
			if field != nil && field.GetName() == name && (field.GetType() == nil || !IsMainTableField(model, field)) {
				return cd.NewError(cd.IllegalParam, fmt.Sprintf("model %s index %s field %s is not stored in the main table", model.GetPkgKey(), s.Name, name))
			}
		}
	}
	return nil
}

// ValidateIndexes is shared by authoring preflight and DDL execution.
func ValidateIndexes(model Model) *cd.Error {
	for _, index := range GetIndexes(model) {
		if err := index.VerifyModel(model); err != nil {
			return err
		}
	}
	return nil
}
