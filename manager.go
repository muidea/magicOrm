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

type manager struct {
	serverConfig *serverConfig

	modelProvider provider.Provider
}

func newManager() *manager {
	return &manager{modelProvider: provider.NewProvider()}
}

func (s *manager) updateServerConfig(cfg *serverConfig) {
	s.serverConfig = cfg
}

func (s *manager) getServerConfig() *serverConfig {
	return s.serverConfig
}

func (s *manager) getProvider() provider.Provider {
	return s.modelProvider
}
