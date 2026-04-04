package helper

import (
	"encoding/json"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/provider/remote"
	"log/slog"
)

func EncodeObject(objPtr *remote.Object) (ret []byte, err *cd.Error) {
	if objPtr == nil {
		err = cd.NewError(cd.IllegalParam, "object is nil")
		return
	}
	byteVal, byteErr := json.Marshal(objPtr)
	if byteErr != nil {
		err = cd.NewError(cd.Unexpected, byteErr.Error())
		slog.Error("EncodeObject Marshal failed", "object", objPtr.GetPkgKey(), "error", byteErr.Error())
		return
	}

	ret = byteVal
	return
}

func DecodeObject(data []byte) (ret *remote.Object, err *cd.Error) {
	if data == nil {
		err = cd.NewError(cd.IllegalParam, "data is nil")
		return
	}
	objPtr := &remote.Object{}
	byteErr := json.Unmarshal(data, objPtr)
	if byteErr != nil {
		err = cd.NewError(cd.Unexpected, byteErr.Error())
		slog.Error("DecodeObject Unmarshal failed", "error", byteErr.Error())
		return
	}

	ret = objPtr
	return
}
