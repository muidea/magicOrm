package database

import (
	"strconv"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
)

// StringKeyLength preserves the legacy width unless a model explicitly declares
// its maximum key length. The bound also applies to reference columns.
func StringKeyLength(spec models.Spec) (int, *cd.Error) {
	if spec == nil || spec.GetConstraints() == nil {
		return 32, nil
	}
	maximum, ok := spec.GetConstraints().Get(models.KeyMax)
	if !ok {
		return 32, nil
	}
	args := maximum.Args()
	if len(args) != 1 {
		return 0, cd.NewError(cd.IllegalParam, "string primary key max requires one integer length")
	}
	size, err := strconv.Atoi(args[0])
	if err != nil || size < 1 || size > 255 {
		return 0, cd.NewError(cd.IllegalParam, "string primary key max must be between 1 and 255")
	}
	return size, nil
}
