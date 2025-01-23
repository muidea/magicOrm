package helper

import (
	"bytes"
	"encoding/json"
	"os"
	"path"
	"strings"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/provider/remote"
)

// GetObject get object
func GetObject(entity any) (*remote.Object, *cd.Result) {
	return remote.GetObject(entity)
}

// GetObjectValue get object value
func GetObjectValue(entity any) (*remote.ObjectValue, *cd.Result) {
	return remote.GetObjectValue(entity)
}

// GetSliceObjectValue get slice object value
func GetSliceObjectValue(sliceEntity any) (*remote.SliceObjectValue, *cd.Result) {
	return remote.GetSliceObjectValue(sliceEntity)
}

func EncodeObject(objPtr *remote.Object) (ret []byte, err *cd.Result) {
	byteVal, byteErr := json.Marshal(objPtr)
	if byteErr != nil {
		err = cd.NewResult(cd.UnExpected, byteErr.Error())
		return
	}

	ret = byteVal
	return
}

func DecodeObject(data []byte) (ret *remote.Object, err *cd.Result) {
	objPtr := &remote.Object{}
	byteErr := json.Unmarshal(data, objPtr)
	if byteErr != nil {
		err = cd.NewResult(cd.UnExpected, byteErr.Error())
		return
	}

	ret = objPtr
	return
}

func SerializeEntity(entity any, destinationPath string) {
	objectPtr, objectErr := GetObject(entity)
	if objectErr != nil {
		log.Errorf("SerializeEntity failed, GetObject error:%s", objectErr.Error())
		return
	}

	byteVal, byteErr := json.Marshal(objectPtr)
	if byteErr != nil {
		log.Errorf("SerializeEntity failed, json.Marshal error:%s", byteErr.Error())
		return
	}
	var byteStream bytes.Buffer
	byteErr = json.Indent(&byteStream, byteVal, "", "\t")
	if byteErr != nil {
		log.Errorf("SerializeEntity failed, json.Indent error:%s", byteErr.Error())
		return
	}

	fileName := strings.ToLower(objectPtr.GetName()) + ".json"
	fileName = path.Join(destinationPath, fileName)
	fileHandle, fileErr := os.OpenFile(fileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if fileErr != nil {
		log.Errorf("SerializeEntity failed, os.Open %s error:%s", fileName, fileErr.Error())
		return
	}
	defer fileHandle.Close()

	writeOK := false
	for {
		_, writeErr := byteStream.WriteTo(fileHandle)
		if writeErr != nil {
			log.Errorf("SerializeEntity failed, fileHandle.Write %s error:%s", fileName, writeErr.Error())
			break
		}

		writeOK = true
		break
	}
	if !writeOK {
		_ = os.Remove(fileName)
		return
	}

	log.Infof("SerializeEntity %s ok!", fileName)
}
