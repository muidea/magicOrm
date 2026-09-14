package database

import cd "github.com/muidea/magicCommon/def"

// SchemaProvisioner is the small database-owner capability used to provision
// an explicitly named schema before an ORM pool is opened.  It is deliberately
// separate from Pool.Initialize: opening a pool must continue to validate a
// pre-existing schema and must not create storage as a side effect.
type SchemaProvisioner interface {
	EnsureSchema() *cd.Error
}
