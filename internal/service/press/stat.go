package press

import (
	"sync"
	"time"
)

type IStat interface {
	RecordSuccess()
	RecordFailure()
	RecordThread()
	RecordTime(t time.Time)
	Snapshot() Snapshot
	Close() error
}

type Snapshot struct {
	Throughput               int64
	ThroughputMean           int64
	ResponseTimeMin          int64
	ResponseTimeMean         int64
	ResponseTimeMax          int64
	TotalFailureRequestCount int64
	TotalRequestCount        int64
	ThreadNum                int64
}

var _ IStat = (*PressStat)(nil)

type PressStat struct {
	Lock                     sync.Mutex
	Once                     sync.Once
	TotalRequestCount        int64     // 请求次数
	TotalStatCount           int64     // 统计次数
	Throughput               int64     // 吞吐量
	TotalSuccessRequestCount int64     // 请求次数-成功
	TotalFailureRequestCount int64     // 请求次数-失败
	MinResponseTime          int64     // 最小响应时间
	MaxResponseTime          int64     // 最大响应时间
	TotalResponseTime        int64     // 总响应时间-均值计算
	ThreadNum                int64     // 现成数
	Closed                   bool      // 关闭
	StartedAt                time.Time // 开始时间
	ClosedAt                 time.Time // 关闭时间
	LastSnapshotAt           time.Time // 最后一次快照时间
}

func (p *PressStat) Snapshot() Snapshot {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	if p.TotalRequestCount < 1 {
		return Snapshot{}
	}
	throughput := int64(float64(p.Throughput) / float64(time.Since(p.LastSnapshotAt).Milliseconds()) * 1000)
	p.Throughput = 0
	p.LastSnapshotAt = time.Now()
	p.TotalStatCount++
	return Snapshot{
		Throughput:               throughput,
		ThroughputMean:           int64(float64(p.TotalRequestCount) / float64(time.Since(p.StartedAt).Milliseconds()) * 1000),
		ResponseTimeMin:          p.MinResponseTime,
		ResponseTimeMean:         int64(float64(p.TotalResponseTime) / float64(p.TotalRequestCount)),
		ResponseTimeMax:          p.MaxResponseTime,
		TotalFailureRequestCount: p.TotalFailureRequestCount,
		TotalRequestCount:        p.TotalRequestCount,
		ThreadNum:                p.ThreadNum,
	}
}

func (p *PressStat) Close() error {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	p.Closed = true
	p.ClosedAt = time.Now()
	return nil
}

func (p *PressStat) RecordSuccess() {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	if p.Closed {
		return
	}
	p.TotalRequestCount++
	p.TotalSuccessRequestCount++
	p.Throughput++
}

func (p *PressStat) RecordFailure() {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	if p.Closed {
		return
	}
	p.TotalRequestCount++
	p.TotalFailureRequestCount++
	p.Throughput++
}

func (p *PressStat) RecordThread() {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	if p.Closed {
		return
	}
	p.ThreadNum++
}

func (p *PressStat) RecordTime(startTime time.Time) {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	if p.Closed {
		return
	}
	responseTime := time.Since(startTime).Milliseconds()
	p.Once.Do(func() {
		p.MinResponseTime = responseTime
		p.StartedAt = time.Now()
		p.LastSnapshotAt = p.StartedAt
	})
	if responseTime < p.MinResponseTime {
		p.MinResponseTime = responseTime
	}
	if responseTime > p.MaxResponseTime {
		p.MaxResponseTime = responseTime
	}
	p.TotalResponseTime += responseTime
}
