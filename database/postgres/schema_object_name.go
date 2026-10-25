package postgres

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/muidea/magicOrm/models"
)

// PostgreSQL indexes (including UNIQUE backing indexes) are schema-scoped.
// Model declarations are table-local. Preserve that contract physically and
// avoid PostgreSQL's silent 63-byte identifier truncation for long names.
func schemaObjectName(table, declaration string) string {
	name := table + "_" + declaration
	// The tuple separator is unambiguous: (A_B,C) and (A,B_C) must differ.
	hash := sha256.Sum256([]byte(table + "\x00" + declaration))
	if len(name) > 46 {
		name = name[:46]
	}
	return name + "_" + hex.EncodeToString(hash[:8])
}

// SchemaObjectName maps a table-local declaration to a physical catalog name.
func (s *Builder) SchemaObjectName(model models.Model, name string) string {
	return schemaObjectName(s.buildCodec.ConstructModelTableName(model), name)
}
