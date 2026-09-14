package orm

import cd "github.com/muidea/magicCommon/def"

// EnsureSchema explicitly provisions one storage schema through the selected
// database adapter. It is intended for deployment/storage owners only. Normal
// AddDatabase calls still reject a missing schema instead of creating one.
func EnsureSchema(dbServer, dbName, schemaName, username, password string) *cd.Error {
	return ensureSchema(NewConfig(dbServer, dbName, schemaName, username, password))
}
