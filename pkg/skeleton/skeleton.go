package skeleton

// KV exposed
type KV interface {
	Set(key string, value uint64) error
	Get(key string) (uint64, bool)
	Push(key string, value uint64) error
	Load(filename string) error
	Save(filename string) error
}
