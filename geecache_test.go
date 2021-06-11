package geecache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

func TestGetterFunc_Get(t *testing.T) {

	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("key")
	got, _ := f.Get("key")

	if !reflect.DeepEqual(got, expect) {
		t.Fatalf("expect %v, got %v", expect, got)
	}

}

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestGet(t *testing.T) {

	loadCounts := make(map[string]int, len(db))

	group := NewGroup("score", 1<<10, GetterFunc(func(key string) ([]byte, error) {

		r, ok := db[key]
		if ok {
			log.Println("[slow db] search key", key)

			_, ok = loadCounts[key]
			if !ok {
				loadCounts[key] = 0
			}

			loadCounts[key] += 1

			return []byte(r), nil
		}

		return nil, fmt.Errorf("%s not exist", key)

	}))

	for k, v := range db {

		vv, err := group.Get(k)
		if err != nil || v != string(vv.b) {
			t.Fatalf("error get key %s, expect %s, got %s", k, v, vv)
		}

		_, err = group.Get(k)
		if err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		}
	}

	vv, err := group.Get("unknown")
	if err == nil {
		t.Fatalf("the value unknown should be empty, but %s got", vv)
	}

}
