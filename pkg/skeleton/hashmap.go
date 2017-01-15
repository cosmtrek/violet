package skeleton

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"sync"

	"github.com/pkg/errors"
)

const (
	tableLen = 0x500
	hash0    = 0
	hashA    = 1
	hashB    = 2
	maxASCII = '\u007f'
)

var (
	cryptTable [tableLen]uint64
	bucketList = []int{17, 37, 79, 163, 331, 673, 1361, 2729, 5471, 10949, 21911, 43853, 87719, 175447, 350899, 701819, 1403641, 2807303, 5614657, 11229331, 22458671, 44917381, 89834777, 179669557, 359339171, 718678369, 1437356741, 2147483647}
)

// HashMap holds the whole data
type HashMap struct {
	buckets   uint64 // bucket length
	length    uint64 // data size
	available bool
	hashtable []hashtable
	entries   []entry
	sync.RWMutex
}

type hashtable struct {
	isOld   bool
	entries []*entry
}

type entry struct {
	Hash0 uint64
	HashA uint64 // first hash
	HashB uint64 // second hash
	Value uint64 // stores index offset
}

// NewHashMap initializes a hashmap
func NewHashMap() *HashMap {
	initCryptTable()
	return &HashMap{
		length:    0,
		entries:   make([]entry, 0),
		available: false,
	}
}

func newEntry(h0, ha, hb, val uint64) entry {
	return entry{
		Hash0: h0,
		HashA: ha,
		HashB: hb,
		Value: val,
	}
}

func (h *HashMap) String() string {
	return fmt.Sprintf("hashmap, buckets: %v, length: %v, available: %v", h.buckets, h.length, h.available)
}

// Load reads data from a file and stores it into hashmap
func (h *HashMap) Load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	h.Lock()
	defer h.Unlock()

	fileLen := int(fileInfo.Size())
	entryCount := fileLen / (8 * 4)
	entries := make([]entry, entryCount)
	h.buckets = calcBuckets(entryCount)

	err = binary.Read(file, binary.LittleEndian, entries)
	if err != nil {
		return err
	}
	h.entries = entries

	h.hashtable = make([]hashtable, h.buckets)
	for i := range h.entries {
		pos := h.entries[i].Hash0 % uint64(h.buckets)
		if !h.hashtable[pos].isOld {
			h.hashtable[pos].entries = make([]*entry, 0)
			h.hashtable[pos].entries = append(h.hashtable[pos].entries, &(h.entries[i]))
			h.hashtable[pos].isOld = true
			continue
		}

		for _, e := range h.hashtable[pos].entries {
			// update entry
			if e.HashA == h.entries[i].HashA && e.HashB == h.entries[i].HashB {
				e.Value = h.entries[i].Value
				continue
			}
		}

		// insert an new entry
		h.hashtable[pos].entries = append(h.hashtable[pos].entries, &(h.entries[i]))
	}
	h.available = true
	return nil
}

// Save persists hashmap's entries
func (h *HashMap) Save(filename string) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	h.Lock()
	buf := new(bytes.Buffer)
	if err = binary.Write(buf, binary.LittleEndian, h.entries); err != nil {
		return err
	}
	h.Unlock()
	if _, err = file.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}

// Push inserts a pair of key and value into hashmap's entries
func (h *HashMap) Push(key string, val uint64) error {
	h0 := hashCode(key, hash0)
	ha := hashCode(key, hashA)
	hb := hashCode(key, hashB)
	e := newEntry(h0, ha, hb, val)
	h.Lock()
	h.entries = append(h.entries, e)
	defer h.Unlock()
	return nil
}

// Set updates entry by key if key exists, otherwise inserts a new pair of key and value
func (h *HashMap) Set(key string, val uint64) error {
	h.Lock()
	defer h.Unlock()

	if !h.available {
		return errors.New("hashmap has not initialized")
	}

	h0 := hashCode(key, hash0)
	ha := hashCode(key, hashA)
	hb := hashCode(key, hashB)
	pos := h0 % uint64(h.buckets)
	e := newEntry(h0, ha, hb, val)

	if !h.hashtable[pos].isOld {
		h.entries = append(h.entries, e)
		h.hashtable[pos].entries = make([]*entry, 0)
		h.hashtable[pos].entries = append(h.hashtable[pos].entries, &e)
		h.hashtable[pos].isOld = true
		return nil
	}

	for _, e := range h.hashtable[pos].entries {
		// update entry
		if e.HashA == ha && e.HashB == hb {
			e.Value = val
			return nil
		}
	}

	// insert new key and value
	h.entries = append(h.entries, e)
	h.hashtable[pos].entries = append(h.hashtable[pos].entries, &e)
	return nil
}

// Get fetches value by key if key is found
func (h *HashMap) Get(key string) (uint64, bool) {
	h0 := hashCode(key, hash0)
	ha := hashCode(key, hashA)
	hb := hashCode(key, hashB)
	pos := h0 % uint64(h.buckets)

	h.RLock()
	defer h.RUnlock()
	if !h.hashtable[pos].isOld {
		return 0, false
	}
	if len(h.hashtable[pos].entries) == 1 {
		return h.hashtable[pos].entries[0].Value, true
	}
	for _, e := range h.hashtable[pos].entries {
		if e.HashA == ha && e.HashB == hb {
			return e.Value, true
		}
	}
	return 0, false
}

func initCryptTable() {
	var seed, idx1, idx2 uint64 = 0x00100001, 0, 0
	i := 0
	for idx1 = 0; idx1 < 0x100; idx1++ {
		for i, idx2 = 0, idx1; i < 5; idx2 += 0x100 {
			seed = (seed*125 + 3) % 0x2aaaab
			a := (seed & 0xffff) << 0x10
			seed = (seed*125 + 3) % 0x2aaaab
			b := seed & 0xffff
			cryptTable[idx2] = a | b
			i++
		}
	}
}

func calcBuckets(maxsize int) uint64 {
	buckets := 0
	if maxsize == 0 {
		return uint64(buckets)
	}
	buckets = findMinBuckets(maxsize)
	for _, size := range bucketList {
		if buckets < size {
			buckets = size
			break
		}
	}
	return uint64(buckets)
}

func findMinBuckets(size int) int {
	buckets := 0
	v := size
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	b := size * 4 / 3
	if b > v {
		buckets = b
	} else {
		buckets = v
	}
	return buckets
}

func hashCode(lpszStr string, dwHashType int) uint64 {
	i, ch := 0, 0
	var seed1, seed2 uint64 = 0x7FED7FED, 0xEEEEEEEE
	var key uint8
	strLen := len(lpszStr)
	for i < strLen {
		key = lpszStr[i]
		ch = int(toUpper(rune(key)))
		i++
		seed1 = cryptTable[(dwHashType<<8)+ch] ^ (seed1 + seed2)
		seed2 = uint64(ch) + seed1 + seed2 + (seed2 << 5) + 3
	}
	return uint64(seed1)
}

func toUpper(r rune) rune {
	if r <= maxASCII {
		if 'a' <= r && r <= 'z' {
			r -= 'a' - 'A'
		}
		return r
	}
	return r
}
