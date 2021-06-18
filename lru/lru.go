package lru

import "container/list"

type Cache struct {
	maxBytes int64
	nbytes   int64

	ll    *list.List
	cache map[string]*list.Element

	OnEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	ele, ok := c.cache[key]
	if !ok {
		return
	}

	c.ll.MoveToBack(ele)

	kv := ele.Value.(*entry)
	value = kv.value

	return
}

func (c *Cache) RemoveOldest() {
	oldestElem := c.ll.Front()

	entry := c.ll.Remove(oldestElem).(*entry)

	delete(c.cache, entry.key)

	c.nbytes -= int64(len(entry.key)) + int64(entry.value.Len())

	if c.OnEvicted != nil {
		c.OnEvicted(entry.key, entry.value)
	}
}

func (c *Cache) Add(key string, value Value) {
	elem, ok := c.cache[key]

	if ok {
		oldEntry := elem.Value.(*entry)

		//elem.Value = value
		oldEntry.value = value
		c.ll.MoveToBack(elem)

		c.nbytes += int64(value.Len()) - int64(oldEntry.value.Len())

	} else {
		newEntry := &entry{key, value}
		elem := c.ll.PushBack(newEntry)

		c.cache[key] = elem

		c.nbytes += int64(len(key)) + int64(value.Len())
	}

	// c.maxBytes 为 0, 代表无限大
	if c.maxBytes == 0 {
		return
	}

	for c.nbytes > c.maxBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
