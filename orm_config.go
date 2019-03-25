package orm

import (
	"github.com/muidea/magicOrm/provider"
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

func newConfig(localProvider bool) *ormConfig {
	if localProvider {
		return &ormConfig{modelProvider: provider.NewLocalProvider()}
	}

	return &ormConfig{modelProvider: provider.NewRemoteProvider()}
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
