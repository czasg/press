package service

import (
	"sync"
	"time"
)

type Stat struct {
	Lock                     sync.Mutex `json:"-"`
	Once                     sync.Once  `json:"-"`
	TotalRequestCount        int64      `json:"totalRequestCount"`        // 请求次数
	TotalStatCount           int64      `json:"totalStatCount"`           // 统计次数
	Throughput               int64      `json:"throughput"`               // 吞吐量
	TotalSuccessRequestCount int64      `json:"totalSuccessRequestCount"` // 请求次数-成功
	TotalFailureRequestCount int64      `json:"totalFailureRequestCount"` // 请求次数-失败
	MinResponseTime          int64      `json:"minResponseTime"`          // 最小响应时间
	MaxResponseTime          int64      `json:"maxResponseTime"`          // 最大响应时间
	TotalResponseTime        int64      `json:"totalResponseTime"`        // 总响应时间-均值计算
	ThreadNum                int64      `json:"threadNum"`                // 现成数
	Closed                   bool       `json:"closed"`                   // 关闭
	StartedAt                time.Time  `json:"startedAt"`                // 开始时间
	ClosedAt                 time.Time  `json:"closedAt"`                 // 关闭时间
	LastSnapshotAt           time.Time  `json:"lastSnapshotAt"`           // 最后一次快照时间
}

func (p *Stat) Snapshot() Snapshot {
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

//func (p *Stat) Start(ctx context.Context) {
//	var handler SnapshotHandler
//	ticker := time.NewTicker(time.Second * time.Duration(p.Step.LogInterval))
//	defer ticker.Stop()
//	for {
//		select {
//		case <-ctx.Done():
//			return
//		case <-ticker.C:
//			handler(ctx, p.Snapshot())
//		}
//	}
//}

func (p *Stat) Close() error {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	p.Closed = true
	p.ClosedAt = time.Now()
	return nil
}

func (p *Stat) RecordSuccess() {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	if p.Closed {
		return
	}
	p.TotalRequestCount++
	p.TotalSuccessRequestCount++
	p.Throughput++
}

func (p *Stat) RecordFailure() {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	if p.Closed {
		return
	}
	p.TotalRequestCount++
	p.TotalFailureRequestCount++
	p.Throughput++
}

func (p *Stat) RecordThread() {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	if p.Closed {
		return
	}
	p.ThreadNum++
}

func (p *Stat) RecordTime(startTime time.Time) {
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
