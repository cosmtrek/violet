package skeleton

import (
	"testing"

	"bytes"
	"encoding/binary"

	"github.com/stretchr/testify/assert"
)

func TestNewHashMap(t *testing.T) {
	assert.NotNil(t, NewHashMap())
}

func TestPush(t *testing.T) {
	hashmap := NewHashMap()
	err := hashmap.Push("key", uint64(1))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(hashmap.entries))
}

func mockedHashmap() *HashMap {
	hashmap := NewHashMap()
	_ = hashmap.Push("key1", uint64(1))
	_ = hashmap.Push("key2", uint64(1))
	_ = hashmap.Push("key3", uint64(1))

	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, hashmap.entries); err != nil {
		panic(err)
	}
	hashmap.buckets = calcBuckets(len(buf.Bytes()))
	entries := make([]entry, 3)
	if err := binary.Read(bytes.NewReader(buf.Bytes()), binary.LittleEndian, entries); err != nil {
		panic(err)
	}
	hashmap.entries = entries
	hashmap.hashtable = make([]hashtable, hashmap.buckets)
	for i := range hashmap.entries {
		pos := hashmap.entries[i].Hash0 % uint64(hashmap.buckets)
		hashmap.hashtable[pos].entries = append(hashmap.hashtable[pos].entries, &(hashmap.entries[i]))
	}
	hashmap.available = true
	return hashmap
}

func TestSetGet(t *testing.T) {
	hashmap := mockedHashmap()
	err := hashmap.Set("key", uint64(1))
	assert.Nil(t, err)
	val, ok := hashmap.Get("key")
	assert.True(t, ok)
	assert.Equal(t, uint64(1), val)
}
