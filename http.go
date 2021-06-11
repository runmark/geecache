package geecache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultBasePath = "/_geecache/"

type HTTPPool struct {
	self     string
	basePath string
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self, defaultBasePath,
	}
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
