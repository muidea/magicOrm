package codec

import (
	"encoding/json"
	"io"
	"strconv"
	"strings"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
)

// decodeStoredCollection reads the declared JSON array representation.
func decodeStoredCollection(text string, typ models.Type) (any, *cd.Error) {
	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, cd.NewError(cd.DataCorrupted, "invalid stored collection JSON")
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, cd.NewError(cd.DataCorrupted, "stored collection has trailing JSON")
	}
	if value == nil {
		return []any(nil), nil
	}
	values, ok := value.([]any)
	if !ok {
		return nil, cd.NewError(cd.DataCorrupted, "stored collection is not an array")
	}
	signed, unsigned := false, false
	switch typ.Elem().GetValue() {
	case models.TypeByteValue, models.TypeSmallIntegerValue, models.TypeInteger32Value, models.TypeBigIntegerValue, models.TypeIntegerValue:
		signed = true
	case models.TypePositiveByteValue, models.TypePositiveSmallIntegerValue, models.TypePositiveInteger32Value, models.TypePositiveBigIntegerValue, models.TypePositiveIntegerValue:
		unsigned = true
	}
	if signed || unsigned {
		for _, item := range values {
			if item == nil && typ.Elem().IsPtrType() {
				continue
			}
			number, ok := item.(json.Number)
			if !ok {
				return nil, cd.NewError(cd.DataCorrupted, "stored integer collection contains a non-integer")
			}
			var err error
			if signed {
				_, err = number.Int64()
			} else {
				_, err = strconv.ParseUint(number.String(), 10, 64)
			}
			if err != nil {
				return nil, cd.NewError(cd.DataCorrupted, "stored integer collection contains an invalid integer")
			}
		}
	}
	// Provider conversion enforces the element's narrower declared range.
	return values, nil
}
