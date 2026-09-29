package mysql

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	cd "github.com/muidea/magicCommon/def"
)

// databaseError preserves portable unique-constraint classification. Other
// failures, including uncertain transaction outcomes, remain unexpected.
func databaseError(err error) *cd.Error {
	if err == nil {
		return nil
	}
	var driverErr *mysql.MySQLError
	if errors.As(err, &driverErr) && driverErr.Number == 1062 {
		return cd.NewError(cd.Duplicated, err.Error())
	}
	return cd.NewError(cd.Unexpected, err.Error())
}
