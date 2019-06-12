package orm

import (
	"sync"

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

	localProviderFlag bool

	ownermodelProviderMap map[string]provider.Provider
	modelProviderLock     sync.RWMutex
}

func newConfig(localProvider bool) *ormConfig {
	return &ormConfig{localProviderFlag: localProvider, ownermodelProviderMap: map[string]provider.Provider{}}
}

func (s *ormConfig) updateServerConfig(cfg *serverConfig) {
	s.serverConfig = cfg
}

func (s *ormConfig) getServerConfig() *serverConfig {
	return s.serverConfig
}

func (s *ormConfig) getProvider(owner string) provider.Provider {
	s.modelProviderLock.Lock()
	defer s.modelProviderLock.Unlock()

	curProvider, ok := s.ownermodelProviderMap[owner]
	if ok {
		return curProvider
	}

	if s.localProviderFlag {
		curProvider = provider.NewLocalProvider()
	} else {
		curProvider = provider.NewRemoteProvider()
	}
	s.ownermodelProviderMap[owner] = curProvider

	return curProvider
}
