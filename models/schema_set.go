package models

import cd "github.com/muidea/magicCommon/def"

// SchemaSetLifecycle is an optional capability for a caller-owned complete
// root set. The caller must freeze the declarations, prove external ownership,
// and serialize writers. It does not grant authority or drop a SQL namespace.
// Existing ORM implementations need not implement this additive capability.
type SchemaSetLifecycle interface {
	InspectSchemaSet(roots []Model, expectedExist bool) (*SchemaReport, *cd.Error)
	// DropSchemaSet verifies every remaining table before any write, drops each
	// shared owned table once, and confirms absence. Missing tables are accepted
	// to resume the same frozen operation; drift is never repaired implicitly.
	DropSchemaSet(roots []Model) (*SchemaReport, *cd.Error)
}
