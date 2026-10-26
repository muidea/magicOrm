package codec

import (
	"fmt"
)

type relationType int

const (
	relationInvalid = 0
	relationHas1v1  = 1
	relationHas1vn  = 2
	relationRef1v1  = 3
	relationRef1vn  = 4
)

func (s relationType) String() string {
	return fmt.Sprintf("%d", s)
}
