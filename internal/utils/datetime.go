package utils

import (
	"reflect"
	"time"

	cd "github.com/muidea/magicCommon/def"
	fu "github.com/muidea/magicCommon/foundation/util"
)

// EncodeUTCDateTime produces the existing timezone-free database representation.
// Such text always denotes UTC, never the process or database session timezone.
func EncodeUTCDateTime(value reflect.Value) (string, *cd.Error) {
	instant, err := ConvertToDateTime(value)
	if err != nil {
		return "", err
	}
	if value = reflect.Indirect(value); value.Kind() == reflect.String && value.String() == "" {
		return "", nil
	}
	return instant.UTC().Format(fu.CSTLayoutWithMillisecond), nil
}

func parseUTCDateTime(text string) (time.Time, error) {
	var lastErr error
	for _, layout := range []string{time.RFC3339Nano, fu.CSTLayoutWithMillisecond, "2006-01-02 15:04:05", "2006-01-02"} {
		instant, err := time.Parse(layout, text)
		if err == nil {
			return instant.UTC(), nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}
