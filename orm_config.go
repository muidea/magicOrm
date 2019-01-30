package orm

import (
	"muidea.com/magicOrm/provider"
)

type serverConfig struct {
	user     string
	password string
	address  string
	dbName   string
}

type ormConfig struct {
	serverConfig *serverConfig

	modelProvider provider.Provider
}

func newConfig() *ormConfig {
	return &ormConfig{modelProvider: provider.NewProvider()}
}

func (s *ormConfig) updateServerConfig(cfg *serverConfig) {
	s.serverConfig = cfg
}

func (s *ormConfig) getServerConfig() *serverConfig {
	return s.serverConfig
}

func (s *ormConfig) getProvider() provider.Provider {
	return s.modelProvider
}
