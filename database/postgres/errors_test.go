package postgres

import (
	"errors"
	"fmt"
	"testing"

	"github.com/lib/pq"
	cd "github.com/muidea/magicCommon/def"
)

func TestDatabaseErrorClassifiesOnlyUniqueViolations(t *testing.T) {
	for _, tc := range []struct {
		err  error
		want int
	}{
		{&pq.Error{Code: "23505"}, int(cd.Duplicated)},
		{fmt.Errorf("wrapped: %w", &pq.Error{Code: "23505"}), int(cd.Duplicated)},
		{&pq.Error{Code: "23503"}, int(cd.Unexpected)},
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
