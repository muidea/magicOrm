package filter

// Filter orm query filter
type Filter interface {
	Add(key string, val interface{})
	Builder() (string, error)
}
