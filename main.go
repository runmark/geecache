package main

import (
	"example.com/mark/geecache/geecache"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func CreateGroup() *geecache.Group {
	return geecache.NewGroup("scores", 2<<10, geecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				//log.Printf("[SlowDB] get key %s value %s\n", key, v)
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		},
	))
}

func startCacheServer(addr string, addrs []string, gee *geecache.Group) {
	peers := geecache.NewHTTPPool(addr)

	peers.Set(addrs...)
	gee.RegisterPeers(peers)

	log.Println("geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(addr string, gee *geecache.Group) {
	http.HandleFunc("/api", func(writer http.ResponseWriter, request *http.Request) {
		values := request.URL.Query()
		key := values.Get("key")

		r, err := gee.Get(key)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("hanle func return %s", r)

		writer.Header().Set("Content-Type", "application/octet-stream")
		writer.Write(r.ByteSlice())
	})

	log.Println("fontend server is running at", addr)

	log.Fatal(http.ListenAndServe(addr[7:], nil))
}

func main() {
	var port int
	var api bool

	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.BoolVar(&api, "api", true, "start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	gee := CreateGroup()
	if api {
		go startAPIServer(apiAddr, gee)
	}

	startCacheServer(addrMap[port], addrs, gee)
}
