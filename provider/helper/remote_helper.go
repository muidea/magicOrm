package helper

import (
	"bytes"
	"encoding/json"
	"os"
	"path"
	"strings"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/provider/remote"
	"log/slog"
)

func EncodeObject(objPtr *remote.Object) (ret []byte, err *cd.Error) {
	byteVal, byteErr := json.Marshal(objPtr)
	if byteErr != nil {
		err = cd.NewError(cd.Unexpected, byteErr.Error())
		return
	}

	ret = byteVal
	return
}

func DecodeObject(data []byte) (ret *remote.Object, err *cd.Error) {
	objPtr := &remote.Object{}
	byteErr := json.Unmarshal(data, objPtr)
	if byteErr != nil {
		err = cd.NewError(cd.Unexpected, byteErr.Error())
		return
	}

	ret = objPtr
	return
}

func SerializeEntity(entity any, destinationPath string) {
	objectPtr, objectErr := GetObject(entity)
	if objectErr != nil {
		slog.Error("message")
		return
	}

	byteVal, byteErr := json.Marshal(objectPtr)
	if byteErr != nil {
		slog.Error("message")
		return
	}
	var byteStream bytes.Buffer
	byteErr = json.Indent(&byteStream, byteVal, "", "\t")
	if byteErr != nil {
		slog.Error("message")
		return
	}

	fileName := strings.ToLower(objectPtr.GetName()) + ".json"
	fileName = path.Join(destinationPath, fileName)
	fileHandle, fileErr := os.OpenFile(fileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if fileErr != nil {
		slog.Error("message")
		return
	}
	defer fileHandle.Close()

	_, writeErr := byteStream.WriteTo(fileHandle)
	if writeErr != nil {
		slog.Error("message")
		_ = os.Remove(fileName)
		return
	}

	slog.Info("info")
}
