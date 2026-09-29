package postgres

import (
	"errors"

	"github.com/lib/pq"
	cd "github.com/muidea/magicCommon/def"
)

// databaseError preserves portable unique-constraint classification. Other
// failures, including uncertain transaction outcomes, remain unexpected.
func databaseError(err error) *cd.Error {
	if err == nil {
		return nil
	}
	var driverErr *pq.Error
	if errors.As(err, &driverErr) && driverErr.Code == "23505" {
		return cd.NewError(cd.Duplicated, err.Error())
	}
	return cd.NewError(cd.Unexpected, err.Error())
}
