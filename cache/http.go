package cache

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"zcache/cache/consistenthash"
	"zcache/cache/pb"
)

/**
 * @author: zurich
 * @Date: 2024/3/15 22:45
 */

const (
	defaultBasePath = "/_zcache/"
	defaultReplicas = 50
)

type HTTPPool struct {
	self       string //本身的host+port
	basePath   string //每个节点的默认前缀
	mu         sync.Mutex
	peers      *consistenthash.Map    //节点
	httpGetter map[string]*HTTPGetter //每个节点的客户端
}

// NewHTTPPool 创建节点http池
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self, //自身的：host + port
		basePath: defaultBasePath,
	}
}

func (h *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", h.self, fmt.Sprintf(format, v...))
}

// Set 设置/更新池中的分布式节点
func (h *HTTPPool) Set(peers ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.peers = consistenthash.New(defaultReplicas, nil)
	h.peers.Add(peers...)
	h.httpGetter = make(map[string]*HTTPGetter, len(peers))
	for _, peer := range peers {
		h.httpGetter[peer] = &HTTPGetter{baseURL: peer + h.basePath}
	}
}

// PickPeer　根据key选取节点
func (h *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if peer := h.peers.Get(key); peer != "" && peer != h.self {
		h.Log("Pick peer %s", peer)
		return h.httpGetter[peer], true
	}
	return nil, false
}
func (h *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, h.basePath) {
		panic("HTTPPool serving excepted path: " + r.URL.Path)
	}
	// /<basePath>/<groupName>/<key>
	parts := strings.SplitN(r.URL.Path[len(h.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupName := parts[0]
	key := parts[1]
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-type", "application/octet-stream")
	if _, err = w.Write(body); err != nil {
		log.Printf("group[%v] failed to response key[%v]\n", groupName, key)
	}
}

var _ PeerGetter = (*HTTPGetter)(nil)

// HTTPGetter 客户端类
type HTTPGetter struct {
	baseURL string
}

func (h *HTTPGetter) Get(req *pb.Request, resp *pb.Response) error {
	uri := fmt.Sprintf("%v%v%v", h.baseURL, url.QueryEscape(req.GetGroup()), url.QueryEscape(req.GetKey()))
	res, err := http.Get(uri)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body failed: %v", err)
	}
	if err = proto.Unmarshal(bytes, resp); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}
	return nil
}
