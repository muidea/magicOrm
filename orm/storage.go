package orm

import "github.com/muidea/magicOrm/models"

// SchemaReport retains the existing public observation contract. Its data
// belongs to models, independently of the database implementation.
type SchemaReport = models.SchemaReport

type databaseConfig struct{ server, name, schema, username, password string }

// NewConfig creates a driver-independent connection description.
func NewConfig(server, name, schema, username, password string) models.DatabaseConfig {
	return databaseConfig{server, name, schema, username, password}
}
func (c databaseConfig) Server() string   { return c.server }
func (c databaseConfig) Database() string { return c.name }
func (c databaseConfig) Schema() string   { return c.schema }
func (c databaseConfig) Username() string { return c.username }
func (c databaseConfig) Password() string { return c.password }
func (c databaseConfig) GetDsn() string {
	return newConfig(c.server, c.name, c.schema, c.username, c.password).GetDsn()
}
