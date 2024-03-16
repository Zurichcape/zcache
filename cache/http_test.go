package cache

import (
	"fmt"
	"log"
	"net/http"
	"testing"
)

/**
 * @author: zurich
 * @Date: 2024/3/15 23:09
 */

func TestHTTPPool(t *testing.T) {
	NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:8888"
	peers := NewHTTPPool(addr)
	t.Log("zcache is running at", addr)
	t.Fatal(http.ListenAndServe(addr, peers))
}
