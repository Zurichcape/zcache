package cache

import (
	"sync"
	"zcache/cache/lru"
)

/**
 * @author: zurich
 * @Date: 2024/3/15 22:06
 */

type cache struct {
	//使用写锁，因为lru内部涉及链表的操作
	mu     sync.Mutex
	lru    *lru.Cache
	cBytes int64
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cBytes, nil)
	}
	c.lru.Put(key, value)
}

func (c *cache) get(key string) (val ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
