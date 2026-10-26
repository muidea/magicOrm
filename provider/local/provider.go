package local

import (
	"reflect"

	"log/slog"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
)

func checkEntityType(entity any) (ret *TypeImpl, err *cd.Error) {
	entityType := reflect.TypeOf(entity)
	typeImplPtr, typeImplErr := NewType(entityType)
	if typeImplErr != nil {
		err = typeImplErr
		slog.Error("checkEntityType NewType failed", "entityType", entityType.String(), "error", err.Error())
		return
	}
	if !models.IsStruct(typeImplPtr.Elem()) {
		err = cd.NewError(cd.IllegalParam, "entity is invalid")
		slog.Error("checkEntityType: not a struct type", "entityType", entityType.String())
		return
	}

	ret = typeImplPtr
	return
}
