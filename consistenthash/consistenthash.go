package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(key []byte) uint32

type Map struct {
	hash     Hash
	replicas int
	keys     []int
	m        map[int]string
}

func NewMap(replicas int, hashFn Hash) *Map {
	m := &Map{
		hash:     hashFn,
		replicas: replicas,
		m:        make(map[int]string),
	}

	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}

	return m
}

func (m *Map) Add(keys ...string) {

	for _, key := range keys {

		for i := 0; i < m.replicas; i++ {
			keyHash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, keyHash)

			m.m[keyHash] = key
		}

	}

	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string {

	if len(m.keys) == 0 {
		return ""
	}

	keyHash := int(m.hash([]byte(key)))
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= keyHash
	})

	vHost := m.keys[idx%len(m.keys)]

	return m.m[vHost]
}
