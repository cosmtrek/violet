package index

const (
	// EQUAL =
	EQUAL = iota
	// LESS <
	LESS
	// GREATER >
	GREATER
)

// Filter on fields
type Filter struct {
	Field string
	Value uint64
	Ftype uint64
}
