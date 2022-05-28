package DawnCache

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	hash     Hash           // 使用的 hash 函数
	replicas int            // 一个真实节点所对应虚拟节点的数量
	keys     []int          // 哈希环，存储所有的节点，有序的
	hashMap  map[int]string // 存储所有节点与真实节点的映射关系
}

func New(replicas int, hash Hash) *Map {
	m := &Map{
		hash:     hash,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	if hash == nil {
		// 默认哈希函数
		m.hash = crc32.ChecksumIEEE
	}

	return m
}

func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ { // 一个真实节点对应 m.replicas 个虚拟节点
			// 计算哈希 = key+编号
			hash := int(m.hash([]byte(key + strconv.Itoa(i))))
			// 加入到哈希环中
			m.keys = append(m.keys, hash)
			// 存储映射
			m.hashMap[hash] = key
		}
	}
	// 将 keys 排序
	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string {
	if len(key) == 0 {
		return ""
	}

	// 计算哈希
	hash := int(m.hash([]byte(key)))
	// 在哈希环上查找节点
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	// 返回节点
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
