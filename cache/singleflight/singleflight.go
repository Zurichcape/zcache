package singleflight

import "sync"

/**
 * @author: zurich
 * @Date: 2024/3/16 11:33
 */

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex // protect follows
	m  map[string]*call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(call)
	//发起请求前加锁
	c.wg.Add(1)
	//请求添加到m中，表示已经有相同的请求
	g.m[key] = c
	g.mu.Unlock()
	//执行请求
	c.val, c.err = fn()
	//请求结束
	c.wg.Done()
	//更新m
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()
	return c.val, c.err
}
