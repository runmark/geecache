package lru

import "testing"

type String string

func (s String) Len() int {
	return len(s)
}

func TestCache_Get(t *testing.T) {

	c := New(int64(0), nil)

	c.Add("key1", String("1234"))

	v, ok := c.Get("key1")

	if !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}

	v, ok = c.Get("key2")
	if ok {
		t.Fatalf("cache miss failed")
	}

}

func TestCache_RemoveOldest(t *testing.T) {
	k1, v1, k2, v2 := "k1", String("v1"), "k2", String("v2")
	maxBytes := len(k1) + len(k2) + v1.Len() + v2.Len()

	c := New(int64(maxBytes), nil)

	c.Add(k1, v1)
	c.Add(k2, v2)
	c.Add("k3", String("v3"))

	if c.Len() != 2 {
		t.Fatalf("removeOldest failed, got: %v", c)
	}

}
