// Package testdatabase admits explicitly disposable integration targets before
// any network access. It is only consumed by integration test entrypoints.
package testdatabase

import (
	"net/url"
	"os"
	"strings"
	"testing"

	mysql "github.com/go-sql-driver/mysql"
)

func RequirePostgres(t *testing.T, dsn string) *url.URL {
	t.Helper()
	if os.Getenv("MAGICORM_INTEGRATION_ALLOW") != "1" {
		t.Skip("set MAGICORM_INTEGRATION_ALLOW=1 for disposable integration databases")
	}
	u, err := url.Parse(dsn)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Host == "" || u.User == nil || !strings.HasPrefix(strings.TrimPrefix(u.Path, "/"), "magicorm_test_") {
		t.Fatal("explicit disposable magicorm_test_ PostgreSQL URL required")
	}
	password, ok := u.User.Password()
	if !ok || password == "" {
		t.Fatal("explicit test credentials required")
	}
	return u
}
func RequireMySQL(t *testing.T, dsn string) {
	t.Helper()
	if os.Getenv("MAGICORM_INTEGRATION_ALLOW") != "1" {
		t.Skip("set MAGICORM_INTEGRATION_ALLOW=1 for disposable integration databases")
	}
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil || cfg.Addr == "" || cfg.User == "" || cfg.Passwd == "" || !strings.HasPrefix(cfg.DBName, "magicorm_test_") {
		t.Fatal("explicit disposable magicorm_test_ MySQL DSN required")
	}
}
