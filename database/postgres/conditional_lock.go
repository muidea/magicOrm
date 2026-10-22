package postgres

// TransactionActive lets ORM row locks reject autocommit usage.
func (s *ConnExecutor) TransactionActive() bool { return s != nil && s.dbTx != nil }
func (s *HostExecutor) TransactionActive() bool { return s != nil && s.dbTx != nil }
