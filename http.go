package DawnCache

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_dawncache/"
	defaultReplicas = 50
)

type HTTPPool struct {
	self        string // 如 http://127.0.0.1:8080
	basePath    string // 节点间通讯地址的前缀，如 http:// 127.0.1:8080/basePath/groupName/key 用于请求数据
	mu          sync.Mutex
	peers       *Map                   // 一致性哈希，根据 key 来选择节点
	httpGetters map[string]*HTTPGetter // 根据 baseURL 选择 HTTPGetter
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log 输出日志信息
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP 处理查询缓存的请求，实现了 http.Handler 接口
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 判断是否有 basePath
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		http.Error(w, "HTTPPool serving unexpected path: "+r.URL.Path, http.StatusBadRequest)
		return
	}

	// 检查是否有 groupName 和 key
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	// 通过 groupName 获取 group
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group:"+groupName, http.StatusBadRequest)
		return
	}

	// 从缓存中获取数据
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 响应客户端
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

// Set 添加节点
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*HTTPGetter)

	for _, peer := range peers {
		p.httpGetters[peer] = &HTTPGetter{basePath: peer + p.basePath}
	}
}

// PickPeer 实现 PeerPicker 接口，用于根据 key 选择节点
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil) // 检查 HTTPPool 是否实现了 PeerPicker 接口

// HTTPGetter 通过 HTTP 远程获取数据
type HTTPGetter struct {
	basePath string
}

// Get 实现了 PeerGetter 接口，用于远程获取源数据
func (h *HTTPGetter) Get(groupName string, key string) ([]byte, error) {
	url := fmt.Sprintf("%s%s/%s", h.basePath, groupName, key)

	res, err := http.Get(url)
	if err != nil {
		// 发送请求失败
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		// 状态码不是 200
		return nil, fmt.Errorf("server status code: %v", res.StatusCode)
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		// 读取数据失败
		return nil, errors.New("read response body failed")
	}

	return data, nil
}

var _ PeerGetter = (*HTTPGetter)(nil) // 检查实现 PeerGetter 接口
