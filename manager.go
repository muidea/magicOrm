package orm

import (
	"muidea.com/magicOrm/model"
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

	modelInfoCache model.Cache

	modelProvider provider.Provider
}

func newManager() *manager {
	cache := model.NewCache()
	return &manager{modelInfoCache: cache, modelProvider: provider.New(cache)}
}

func (s *manager) updateServerConfig(cfg *serverConfig) {
	s.serverConfig = cfg
}

func (s *manager) getServerConfig() *serverConfig {
	return s.serverConfig
}

func (s *manager) getCache() model.Cache {
	return s.modelInfoCache
}

func (s *manager) getProvider() provider.Provider {
	return s.modelProvider
}
