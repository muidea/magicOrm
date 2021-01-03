package helper

import (
	"github.com/muidea/magicOrm/model"
)

// encodeStructValue get struct value str
func (s *impl) encodeStructValue(vVal model.Value, tType model.Type) (ret string, err error) {
	vModel, vErr := s.getValueModel(vVal, tType)
	if vErr != nil {
		err = vErr
		return
	}

	pkField := vModel.GetPrimaryField()
	ret, err = s.Encode(pkField.GetValue(), pkField.GetType())
	return
}

// decodeStructValue decode struct from string
func (s *impl) decodeStructValue(val string, tType model.Type) (ret model.Value, err error) {
	tVal := tType.Interface(nil)
	vModel, vErr := s.getValueModel(tVal, tType)
	if vErr != nil {
		err = vErr
		return
	}

	pkField := vModel.GetPrimaryField()
	fVal, fErr := s.Decode(val, pkField.GetType())
	if fErr != nil {
		err = fErr
		return
	}
	err = pkField.SetValue(fVal)
	if err != nil {
		return
	}

	if tType.IsPtrType() {
		tVal = tVal.Addr()
	}

	ret = tVal
	return
}
