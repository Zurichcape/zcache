package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	zcache "zcache/cache"
)

/**
 * @author: zurich
 * @Date: 2024/3/15 21:59
 */

var db = map[string]string{
	"zurich": "100",
	"alfred": "92",
	"tomcat": "79",
}

func createGroup() *zcache.Group {
	return zcache.NewGroup("scores", 2<<10, zcache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, cache *zcache.Group) {
	peers := zcache.NewHTTPPool(addr)
	peers.Set(addrs...)
	cache.RegisterPeers(peers)
	log.Println("zcache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, cache *zcache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := cache.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())

		}))
	log.Println("frontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))

}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Zcache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
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

	cache := createGroup()
	if api {
		go startAPIServer(apiAddr, cache)
	}
	startCacheServer(addrMap[port], addrs, cache)
}
