package etcd

import (
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

type HashRing struct {
	mu        sync.RWMutex
	vNodes    int                 // 虚拟节点数，固定200
	keys      []uint32            // 排序后的哈希环
	hashMap   map[uint32]string   // 哈希值 => 真实worker
	workNodes map[string]struct{} // 当前在线节点集合
}

func NewHashRing() *HashRing {
	return &HashRing{
		vNodes:    200,
		hashMap:   make(map[uint32]string),
		workNodes: make(map[string]struct{}),
	}
}

// AddWorker 添加一个Worker
func (h *HashRing) AddWorker(addr string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.workNodes[addr]; ok {
		return
	}
	h.workNodes[addr] = struct{}{}

	// 创建 200 个虚拟节点
	for i := 0; i < h.vNodes; i++ {
		key := addr + "#" + strconv.Itoa(i)
		hash := crc32.ChecksumIEEE([]byte(key))
		h.keys = append(h.keys, hash)
		h.hashMap[hash] = addr
	}

	// 排序环
	sort.Slice(h.keys, func(i, j int) bool {
		return h.keys[i] < h.keys[j]
	})
}

func (h *HashRing) HasKey(addr string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.workNodes[addr]
	return ok
}

// RemoveWorker 删除一个Worker
func (h *HashRing) RemoveWorker(addr string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.workNodes[addr]; !ok {
		return
	}
	delete(h.workNodes, addr)

	// 删除所有虚拟节点
	for i := 0; i < h.vNodes; i++ {
		key := addr + "#" + strconv.Itoa(i)
		hash := crc32.ChecksumIEEE([]byte(key))
		delete(h.hashMap, hash)
	}

	// 重建环
	h.keys = make([]uint32, 0, len(h.hashMap))
	for k := range h.hashMap {
		h.keys = append(h.keys, k)
	}
	sort.Slice(h.keys, func(i, j int) bool {
		return h.keys[i] < h.keys[j]
	})
}

func (h *HashRing) GetWorkerByKey(key string) (string, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.keys) == 0 {
		return "", false
	}

	hash := crc32.ChecksumIEEE([]byte(key))

	// 二分查找
	idx := sort.Search(len(h.keys), func(i int) bool {
		return h.keys[i] >= hash
	})

	// 回到环头
	if idx == len(h.keys) {
		idx = 0
	}

	return h.hashMap[h.keys[idx]], true
}

func (h *HashRing) GetWorkerByRand() (string, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.keys) == 0 {
		return "", false
	}
	for node := range h.workNodes {
		return node, true
	}
	return "", false
}
