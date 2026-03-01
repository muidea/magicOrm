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

func SerializeEntity(entity any, destinationPath string) {
	objectPtr, objectErr := GetObject(entity)
	if objectErr != nil {
		slog.Error("SerializeEntity GetObject failed", "error", objectErr.Error())
		return
	}

	byteVal, byteErr := json.Marshal(objectPtr)
	if byteErr != nil {
		slog.Error("SerializeEntity Marshal failed", "object", objectPtr.GetPkgKey(), "error", byteErr.Error())
		return
	}
	var byteStream bytes.Buffer
	byteErr = json.Indent(&byteStream, byteVal, "", "\t")
	if byteErr != nil {
		slog.Error("SerializeEntity Indent failed", "object", objectPtr.GetPkgKey(), "error", byteErr.Error())
		return
	}

	fileName := strings.ToLower(objectPtr.GetName()) + ".json"
	fileName = path.Join(destinationPath, fileName)
	fileHandle, fileErr := os.OpenFile(fileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if fileErr != nil {
		slog.Error("SerializeEntity OpenFile failed", "path", fileName, "error", fileErr.Error())
		return
	}
	defer fileHandle.Close()

	_, writeErr := byteStream.WriteTo(fileHandle)
	if writeErr != nil {
		slog.Error("SerializeEntity WriteTo failed", "path", fileName, "error", writeErr.Error())
		_ = os.Remove(fileName)
		return
	}

	slog.Info("SerializeEntity done", "object", objectPtr.GetPkgKey(), "path", fileName)
}
