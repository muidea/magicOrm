package postgres

import (
	"errors"

	"github.com/lib/pq"
	cd "github.com/muidea/magicCommon/def"
)

// databaseError classifies confirmed constraint/transaction conflicts without
// retrying them. Connection loss and unknown outcomes remain unexpected; only
// the operation owner can decide what a safe recovery query means.
func databaseError(err error) *cd.Error {
	if err == nil {
		return nil
	}
	var driverErr *pq.Error
	if errors.As(err, &driverErr) {
		switch driverErr.Code {
		case "23505":
			return cd.NewError(cd.Duplicated, err.Error())
		case "40001", "40P01":
			return cd.NewError(cd.VersionConflict, err.Error())
		}
	}
	return cd.NewError(cd.Unexpected, err.Error())
}
