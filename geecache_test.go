package geecache

import (
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
