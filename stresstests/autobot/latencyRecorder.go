package autobot

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type phaseState struct {
	State int
}

type LatencyRecorder struct {
	mu        sync.Mutex
	latencies []float64 // 存放所有延迟数据，单位毫秒
	botCount  int
}

func NewLatencyRecorder(bc int) *LatencyRecorder {
	return &LatencyRecorder{botCount: bc}
}

func (r *LatencyRecorder) Record(latency time.Duration) {
	r.mu.Lock()
	r.latencies = append(r.latencies, float64(latency.Microseconds())/1000.0) // 转成毫秒
	r.mu.Unlock()
}

// 计算并格式化输出报告
func (r *LatencyRecorder) Report() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.latencies) == 0 {
		return "No latency data"
	}

	// 排序以计算p99
	sorted := make([]float64, len(r.latencies))
	copy(sorted, r.latencies)
	sort.Float64s(sorted)

	// 累加计算平均值
	sum := 0.0
	for _, v := range sorted {
		sum += v
	}
	avg := sum / float64(len(sorted))

	// 计算p99延迟
	p99Index := int(float64(len(sorted)) * 0.99)
	p99 := sorted[p99Index]

	// 最大延迟
	maxLatency := sorted[len(sorted)-1]

	return fmt.Sprintf(
		"样本数: %d, 平均延迟: %.2f ms, P99延迟: %.2f ms, 最大延迟: %.2f ms",
		len(sorted), avg, p99, maxLatency,
	)
}
