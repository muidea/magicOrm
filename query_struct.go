package orm

import (
	"fmt"
	"reflect"

	"muidea.com/magicOrm/model"
)

type structValue struct {
	value          reflect.Value
	modelInfoCache model.StructInfoCache
}

func (s *structValue) String() (str string, err error) {
	info, infoErr := model.GetStructValue(s.value, s.modelInfoCache)
	if infoErr != nil {
		err = infoErr
		return
	}

	pk := info.GetPrimaryField()
	if pk == nil {
		err = fmt.Errorf("GetPrimaryField faield, no defined pk")
		return
	}

	pfv := pk.GetFieldValue()
	if pfv == nil {
		err = fmt.Errorf("GetFieldValue faield, value is nil")
		return
	}

	val, valErr := pfv.GetValueStr()
	if valErr != nil {
		err = valErr
		return
	}
	str = val

	return
}
