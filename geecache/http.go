package geecache

import (
	"example.com/mark/geecache/consistenthash"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

type HTTPPool struct {
	self        string
	basePath    string
	mu          sync.Mutex
	peers       *consistenthash.Map
	httpGetters map[string]*httpGetter
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self: self, basePath: defaultBasePath,
	}
}

func (pool *HTTPPool) Set(peers ...string) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	pool.peers = consistenthash.NewMap(defaultReplicas, nil)
	pool.peers.Add(peers...)

	pool.httpGetters = make(map[string]*httpGetter, len(peers))

	for _, peer := range peers {
		pool.httpGetters[peer] = &httpGetter{baseURL: peer + pool.basePath}
	}
}

var _ PeerPicker = (*HTTPPool)(nil)

func (pool *HTTPPool) PickPeer(key string) (peer PeerGetter, ok bool) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	peerId := pool.peers.Get(key)
	if peerId == "" {
		pool.Log("cannot find peer by key %s", key)
		return nil, false
	}

	peer, ok = pool.httpGetters[peerId]
	return
}

func (pool *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[server %s] %s", pool.self, fmt.Sprintf(format, v...))
}

func (pool *HTTPPool) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	// p's format: "/_geecache/<groupName>/<cacheName>/..."
	if !strings.HasPrefix(r.URL.Path, pool.basePath) {
		pool.Log("basePath: %s, requestPath: %s", pool.basePath, r.URL.Path)
		panic("HTTPPool serving error path: " + r.URL.Path)
	}

	pool.Log("%s %s", r.Method, r.URL.Path)

	gc := strings.SplitN(r.URL.Path[len(defaultBasePath):], "/", 2)

	if len(gc) != 2 {
		http.Error(rw, "bad request", http.StatusBadRequest)
		return
	}

	groupName, cacheName := gc[0], gc[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(rw, "no such group: "+group.name, http.StatusNotFound)
		return
	}

	view, err := group.Get(cacheName)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/octet-stream")
	rw.Write(view.b)
}

type httpGetter struct {
	baseURL string
}

var _ PeerGetter = (*httpGetter)(nil)

func (h *httpGetter) Get(group string, key string) (bytes []byte, err error) {
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(group), url.QueryEscape(key))
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %v", resp.Status)
	}

	_, err = resp.Body.Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return
}

