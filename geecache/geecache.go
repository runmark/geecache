package geecache

import (
	"errors"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

// 接口型函数
type GetterFunc func(key string) ([]byte, error)

func (gf GetterFunc) Get(key string) ([]byte, error) {
	return gf(key)
}

type Group struct {
	name      string
	getter    Getter
	mainCache cache
}

var (
	mu sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, maxBytes int64, getter Getter) (group *Group) {

	if getter == nil {
		panic("getter is nil")
	}

	group = &Group{
		name, getter, cache{cacheBytes: maxBytes},
	}

	mu.Lock()
	groups[name] = group
	mu.Unlock()

	return
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()

	return groups[name]
}

func (g *Group) Get(key string) (r ByteView, err error) {

	if key == "" {
		return ByteView{}, errors.New("key is required")
	}

	// why can it be called?
	r, ok := g.mainCache.get(key)
	if ok {
		return r, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (ByteView, error) {
	return g.GetFromLocal(key)
}

func (g *Group) GetFromLocal(key string) (bv ByteView, err error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	// prevent changing cache value unintentionally
	bv = ByteView{cloneBytes(bytes)}

	g.populateCache(key, bv)

	return bv, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

