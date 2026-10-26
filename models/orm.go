package models

import cd "github.com/muidea/magicCommon/def"

// ORM is the stable public query/write contract exposed by magicOrm.
//
// Query is the single-object query entrypoint driven by a model instance.
// BatchQuery is the multi-object query entrypoint driven by a filter.
// magicOrm intentionally keeps only these two public query shapes to avoid
// overlapping single-row semantics in the ORM layer.
type ORM interface {
	Create(entity Model) *cd.Error
	Drop(entity Model) *cd.Error
	// Reconcile applies only safe, additive differences between two registered
	// model declarations. Destructive changes are rejected before any SQL runs.
	Reconcile(previous Model, current Model) *cd.Error
	// InspectSchema reads physical catalogs; false expects all owned tables to
	// be absent after Drop. It never applies or repairs DDL.
	InspectSchema(entity Model, expectedExist bool) (*SchemaReport, *cd.Error)
	// RepairSchema verifies catalogs, applies only missing original steps, then
	// independently confirms the target. Callers own authorization/serialization.
	RepairSchema(previous Model, desired Model) (*SchemaReport, *cd.Error)
	Insert(entity Model) (Model, *cd.Error)
	Update(entity Model) (Model, *cd.Error)
	// UpdateWithFilter performs a single-table conditional update and returns
	// the database affected-row count. It is suitable for compare-and-set
	// transitions and participates in an explicitly started transaction.
	UpdateWithFilter(entity Model, filter Filter) (int64, *cd.Error)
	Delete(entity Model) (Model, *cd.Error)
	Query(entity Model) (Model, *cd.Error)
	Count(filter Filter) (int64, *cd.Error)
	BatchQuery(filter Filter) ([]Model, *cd.Error)
	BeginTransaction() *cd.Error
	CommitTransaction() *cd.Error
	RollbackTransaction() *cd.Error
	Release()
}
