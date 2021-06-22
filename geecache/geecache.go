package geecache

import (
	"errors"
	"example.com/mark/geecache/geecachepb/geecachepb"
	"example.com/mark/geecache/singleflight"
	"log"
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
	peers     PeerPicker
	loader    *singleflight.Group
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, maxBytes int64, getter Getter) (group *Group) {

	if getter == nil {
		panic("getter is nil")
	}

	group = &Group{
		name: name, getter: getter, mainCache: cache{cacheBytes: maxBytes},
		loader: &singleflight.Group{},
	}

	mu.Lock()
	groups[name] = group
	mu.Unlock()

	return
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
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

func (g *Group) load(key string) (value ByteView, err error) {

	viewi, err := g.loader.Do(key, func() (interface{}, error) {

		if g.peers != nil {
			peer, ok := g.peers.PickPeer(key)
			if ok {
				value, err := g.GetFromPeer(peer, key)
				if err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}

		return g.GetFromLocal(key)
	})

	if err == nil {
		return viewi.(ByteView), nil
	}

	return
}

func (g *Group) GetFromPeer(peer PeerGetter, key string) (bv ByteView, err error) {
	req := &geecachepb.Request{Group: g.name, Key: key}
	res := &geecachepb.Response{}

	err = peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}

	return ByteView{b: res.Value}, nil
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
