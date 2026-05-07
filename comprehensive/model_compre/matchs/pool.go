// match/pool.go
package matchs

import (
	"container/heap"
	"sync"
	"time"
)

// 优先队列（按进入时间排序，用于超时扩圈）
type PriorityQueue []*WaitingInfo

func (pq PriorityQueue) Len() int            { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool  { return pq[i].EnterTime.Before(pq[j].EnterTime) }
func (pq PriorityQueue) Swap(i, j int)       { pq[i], pq[j] = pq[j], pq[i] }
func (pq *PriorityQueue) Push(x interface{}) { *pq = append(*pq, x.(*WaitingInfo)) }
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	x := old[n-1]
	*pq = old[0 : n-1]
	return x
}

// 本地匹配池（纯内存）
type LocalMatchPool struct {
	mu sync.RWMutex

	// 匹配者详情
	waiter map[string]*WaitingInfo

	// 分数索引（有序列表，用于范围查询）
	// 使用双向链表 + 跳表的思想，这里简化用数组 + 排序
	ratingIndex []WaitingItem

	// 时间优先队列（用于扩圈扫描）
	timeQueue PriorityQueue

	// 是否已排序（懒排序优化）
	dirty bool
}

func NewLocalMatchPool() *LocalMatchPool {
	return &LocalMatchPool{
		waiter:      make(map[string]*WaitingInfo),
		ratingIndex: make([]WaitingItem, 0),
		timeQueue:   make(PriorityQueue, 0),
		dirty:       false,
	}
}

// Add 添加匹配队伍到匹配池
func (p *LocalMatchPool) Add(waiter *WaitingInfo) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.waiter[waiter.TeamID] = waiter
	p.ratingIndex = append(p.ratingIndex, waiter.ToWaitingItem())
	heap.Push(&p.timeQueue, waiter)
	p.dirty = true
}

// 移除匹配者
func (p *LocalMatchPool) Remove(waiterId string) *WaitingInfo {
	p.mu.Lock()
	defer p.mu.Unlock()

	player, ok := p.waiter[waiterId]
	if !ok {
		return nil
	}
	delete(p.waiter, waiterId)

	// 标记删除，后续整理
	p.dirty = true
	return player
}

// 获取匹配者信息
func (p *LocalMatchPool) Get(waiterId string) *WaitingInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.waiter[waiterId]
}

func (p *LocalMatchPool) GetAllWaiters() []*WaitingInfo {
	result := make([]*WaitingInfo, 0, len(p.waiter))
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, waiter := range p.waiter {
		result = append(result, waiter)
	}
	return result
}

// 获取所有等待匹配者数量
func (p *LocalMatchPool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.waiter)
}

// 按分数范围查找候选匹配者
func (p *LocalMatchPool) FindByRatingRange(minRating, maxRating float64, excludeID string, limit int) []WaitingItem {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// 懒排序：只有在 dirty 时才排序
	if p.dirty {
		p.sortLocked()
	}

	result := make([]WaitingItem, 0, limit)
	for _, item := range p.ratingIndex {
		if item.Score < minRating {
			continue
		}
		if item.Score > maxRating {
			break
		}
		if item.TeamID != excludeID {
			result = append(result, item)
			if len(result) >= limit {
				break
			}
		}
	}
	return result
}

// 获取需要扩圈的匹配者（进入时间超过当前扩圈等级的最大等待时间）
func (p *LocalMatchPool) GetPlayersNeedExpand(now time.Time) []*WaitingInfo {
	p.mu.Lock()
	defer p.mu.Unlock()

	result := make([]*WaitingInfo, 0)

	// 复制一份优先队列（不破坏原队列）
	temp := make(PriorityQueue, len(p.timeQueue))
	copy(temp, p.timeQueue)
	heap.Init(&temp)

	for temp.Len() > 0 {
		waiter := heap.Pop(&temp).(*WaitingInfo)
		expandCfg := DefaultExpandConfig[waiter.ExpandLevel]
		waitSeconds := int(now.Sub(waiter.EnterTime).Seconds())

		if waitSeconds >= expandCfg.MaxWait {
			// 需要扩圈
			result = append(result, waiter)
		} else {
			// 队列是按时间排序的，后面的等待时间更短，不需要继续
			break
		}
	}
	return result
}

// 更新匹配者的扩圈等级
func (p *LocalMatchPool) UpdateExpandLevel(waiterId string, newLevel int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	waiter, ok := p.waiter[waiterId]
	if !ok {
		return
	}
	waiter.ExpandLevel = newLevel
	// 更新时间优先队列中的引用（引用类型，无需额外操作）
}

// 内部排序
func (p *LocalMatchPool) sortLocked() {
	// 按分数排序
	// 使用稳定排序，相同分数按时间（这里简化，按playerID保证确定性）
	for i := 0; i < len(p.ratingIndex)-1; i++ {
		for j := i + 1; j < len(p.ratingIndex); j++ {
			if p.ratingIndex[i].Score > p.ratingIndex[j].Score {
				p.ratingIndex[i], p.ratingIndex[j] = p.ratingIndex[j], p.ratingIndex[i]
			}
		}
	}

	// 清理已删除的匹配者
	valid := make([]WaitingItem, 0, len(p.ratingIndex))
	for _, item := range p.ratingIndex {
		if _, ok := p.waiter[item.TeamID]; ok {
			valid = append(valid, item)
		}
	}
	p.ratingIndex = valid

	// 重建时间队列
	newQueue := make(PriorityQueue, 0, len(p.waiter))
	for _, player := range p.waiter {
		newQueue = append(newQueue, player)
	}
	heap.Init(&newQueue)
	p.timeQueue = newQueue

	p.dirty = false
}
