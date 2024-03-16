package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

/**
 * @author: zurich
 * @Date: 2024/3/15 21:31
 */

type Hash func(data []byte) uint32

// Map 包含所有的hash后的key
type Map struct {
	hash     Hash           //hash函数
	replicas int            //虚拟节点数量
	keys     []int          //哈希环
	hashMap  map[int]string //虚拟节点与真实节点映射表
}

// New 创建Map实例
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

func (m *Map) Add(keys ...string) {
	//真实节点
	for _, key := range keys {
		//每个真实节点创建虚拟节点
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))
	//二分查找
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	//待查找的key可能比所有key都大，查找结果可能为len(keys)
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
