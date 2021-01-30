package orm

type serverConfig struct {
	user     string
	password string
	address  string
	dbName   string
}

type ormConfig struct {
	serverConfig *serverConfig
}

func newConfig() *ormConfig {
	return &ormConfig{}
}

func (s *ormConfig) updateServerConfig(cfg *serverConfig) {
	s.serverConfig = cfg
}

func (s *ormConfig) getServerConfig() *serverConfig {
	return s.serverConfig
}
