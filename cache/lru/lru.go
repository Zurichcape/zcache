package lru

import "container/list"

/**
 * @author: zurich
 * @Date: 2024/3/15 21:31
 */

type Cache struct {
	maxBytes int64
	curBytes int64
	ll       *list.List
	cache    map[string]*list.Element
	//key被置换时触发
	onEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

// New 实例化
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		onEvicted: onEvicted,
	}
}

// Get 查找key的value
func (c *Cache) Get(key string) (val Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		v := ele.Value.(*entry)
		return v.value, true
	}
	return
}

// RemoveOldest 缓存淘汰
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		val := ele.Value.(*entry)
		delete(c.cache, val.key)
		c.curBytes -= int64(len(val.key)) + int64(val.value.Len())
		if c.onEvicted != nil {
			c.onEvicted(val.key, val.value)
		}
	}
}

// Put 新增/修改某个k-v
func (c *Cache) Put(key string, val Value) {
	//修改
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		v := ele.Value.(*entry)
		c.curBytes += int64(val.Len()) - int64(v.value.Len())
		v.value = val
	} else {
		//新增
		ele := c.ll.PushFront(&entry{key: key, value: val})
		c.cache[key] = ele
		c.curBytes += int64(len(key)) + int64(val.Len())
	}
	//淘汰
	for c.maxBytes > 0 && c.curBytes > c.maxBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
