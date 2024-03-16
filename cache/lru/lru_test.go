package lru

import (
	"fmt"
	"testing"
)

/**
 * @author: zurich
 * @Date: 2024/3/15 21:54
 */
type String string

func (d String) Len() int {
	return len(d)
}

func TestGet(t *testing.T) {
	lru := New(int64(0), nil)
	lru.Put("key1", String("1234"))
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "v3"
	cap := len(k1 + k2 + v1 + v2)
	onEvicted := func(key string, val Value) {
		fmt.Println("触发了函数", key, val)
	}
	lru := New(int64(cap), onEvicted)
	lru.Put(k1, String(v1))
	lru.Put(k2, String(v2))
	lru.Put(k3, String(v3))

	if _, ok := lru.Get("key1"); ok || lru.Len() != 2 {
		t.Fatalf("Removeoldest key1 failed")
	}
}
