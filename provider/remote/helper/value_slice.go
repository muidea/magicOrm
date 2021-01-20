package helper

import (
	"encoding/json"

	"github.com/muidea/magicOrm/model"
)

// encodeSliceValue get slice value str
func (s *impl) encodeSliceValue(vVal model.Value, tType model.Type) (ret string, err error) {
	vals, valErr := s.elemDependValue(vVal)
	if valErr != nil {
		err = valErr
		return
	}
	items := []string{}
	for _, val := range vals {
		strVal, strErr := s.Encode(val, tType.Elem())
		if strErr != nil {
			err = strErr
			return
		}

		items = append(items, strVal)
	}

	data, dataErr := json.Marshal(items)
	if dataErr != nil {
		err = dataErr
		return
	}

	ret = string(data)
	return
}

// decodeSliceValue decode slice from string
func (s *impl) decodeSliceValue(val string, tType model.Type) (ret model.Value, err error) {
	items := []string{}
	err = json.Unmarshal([]byte(val), &items)
	if err != nil {
		return
	}

	sliceVal := []interface{}{}
	for idx := range items {
		itemVal, itemErr := s.Decode(items[idx], tType.Elem())
		if itemErr != nil {
			err = itemErr
			return
		}

		sliceVal = append(sliceVal, itemVal.Get())
	}

	tVal, _ := tType.Interface(sliceVal)
	if tType.IsPtrType() {
		tVal = tVal.Addr()
	}

	ret = tVal
	return
}
