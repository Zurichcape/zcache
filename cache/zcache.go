package cache

import (
	"errors"
	"log"
	"sync"
	"zcache/cache/pb"
	"zcache/cache/singleflight"
)

/**
 * @author: zurich
 * @Date: 2024/3/15 22:18
 */

type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 函数式接口
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string              //缓存命名
	getter    Getter              //缓存未命中的回调
	mainCache cache               //分布式缓存
	peers     PeerPicker          //节点选择
	loader    *singleflight.Group //保证并发访问时，相同请求只有一个请求会被处理
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if peers == nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) GetFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{buf: res.Value}, nil
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, errors.New("key is required")
	}
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[zcache] hit!")
		return v, nil
	}
	//未命中则执行回调从其他节点获取
	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	val, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if v, err := g.GetFromPeer(peer, key); err == nil {
					return v, nil
				}
				log.Println("[zcache] failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})
	if err == nil {
		return val.(ByteView), nil
	}
	return
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{buf: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cBytes: cBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	g := groups[name]
	return g
}
