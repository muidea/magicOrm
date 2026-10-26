package mysql

// TransactionActive lets ORM row locks reject autocommit usage.
func (s *ConnExecutor) TransactionActive() bool {
	return s != nil && !s.released && !s.dbRollbackOnly && s.dbTx != nil
}
func (s *HostExecutor) TransactionActive() bool {
	return s != nil && !s.released && !s.dbRollbackOnly && s.dbTx != nil
}
