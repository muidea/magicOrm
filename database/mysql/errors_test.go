package mysql

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"
	cd "github.com/muidea/magicCommon/def"
)

func TestDatabaseErrorClassifiesOnlyUniqueViolations(t *testing.T) {
	for _, tc := range []struct {
		err  error
		want int
	}{
		{&mysql.MySQLError{Number: 1062}, int(cd.Duplicated)},
		{fmt.Errorf("wrapped: %w", &mysql.MySQLError{Number: 1062}), int(cd.Duplicated)},
		{&mysql.MySQLError{Number: 1213}, int(cd.Unexpected)},
		{errors.New("connection lost after write"), int(cd.Unexpected)},
	} {
		got := databaseError(tc.err)
		if got == nil || int(got.Code) != tc.want {
			t.Fatalf("error=%v want code=%d", got, tc.want)
		}
	}
	if databaseError(nil) != nil {
		t.Fatal("nil error changed")
	}
}
