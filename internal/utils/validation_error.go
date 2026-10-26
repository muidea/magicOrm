package utils

import (
	"errors"
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
)

// ConstraintErrorDetails preserves custom validator reasons as well as the
// directive identity supplied by the built-in validator. It never formats val.
func ConstraintErrorDetails(err error) (key models.Key, message string) {
	var violation *models.ConstraintViolation
	if errors.As(err, &violation) {
		return violation.Key, fmt.Sprintf("constraint '%s': %s", violation.Key, err.Error())
	}
	return "", err.Error()
}

func FieldValidationError(field string, err error) *cd.Error {
	_, message := ConstraintErrorDetails(err)
	return cd.NewError(cd.IllegalParam, fmt.Sprintf("field '%s': %s", field, message))
}
