package index

const (
	// EQUAL =
	EQUAL = iota
	// Less <
	LESS
	// GREATER >
	GREATER
)

// Filter
type Filter struct {
	Field string
	Value uint64
	Ftype uint64
}
