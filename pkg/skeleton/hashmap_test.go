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

func mockedHashmap() *HashMap {
	hashmap := NewHashMap()
	_ = hashmap.Push("key1", uint64(0))
	_ = hashmap.Push("key2", uint64(1))
	_ = hashmap.Push("key3", uint64(2))

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
		hashmap.hashtable[pos].entries = make([]*entry, 0)
		hashmap.hashtable[pos].entries = append(hashmap.hashtable[pos].entries, &(hashmap.entries[i]))
		hashmap.hashtable[pos].isOld = true
	}
	hashmap.available = true
	return hashmap
}

func TestHashmap_Push_Set_Get(t *testing.T) {
	hashmap := mockedHashmap()
	val, ok := hashmap.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, uint64(0), val)

	err := hashmap.Set("key", uint64(1))
	assert.Nil(t, err)
	val2, ok := hashmap.Get("key")
	assert.True(t, ok)
	assert.Equal(t, uint64(1), val2)
}
