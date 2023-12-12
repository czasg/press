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
	RecordThroughput(t time.Time)
	Snapshot(t time.Time) Snapshot
	Close() error
}

type Snapshot struct {
	Throughput               int64
	ThroughputMean           int64
	ResponseTimeMin          int64
	ResponseTimeMax          int64
	ResponseTimeMean         int64
	TotalFailureRequestCount int64
	TotalRequestCount        int64
	ThreadNum                int64
}

var _ IStat = (*PressStat)(nil)

type PressStat struct {
	Lock                        sync.Mutex
	Once                        sync.Once
	TotalRequestCount           int64     // 请求次数
	TotalStatCount              int64     // 统计次数
	Throughput                  int64     // 吞吐量
	ThroughputLastCalculateTime time.Time // 吞吐量最后计算时间
	TotalSuccessRequestCount    int64     // 请求次数-成功
	TotalFailureRequestCount    int64     // 请求次数-失败
	MinResponseTime             int64     // 最小响应时间
	MaxResponseTime             int64     // 最大响应时间
	TotalResponseTime           int64     // 总响应时间-均值计算
	ThreadNum                   int64     // 现成数
	Closed                      bool      // 关闭
}

func (p *PressStat) Snapshot(t time.Time) Snapshot {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	p.TotalStatCount++
	return Snapshot{
		Throughput:               p.Throughput / time.Since(t).Milliseconds() * 1000,
		ThroughputMean:           p.TotalSuccessRequestCount / p.TotalStatCount,
		ResponseTimeMin:          p.MinResponseTime,
		ResponseTimeMax:          p.MaxResponseTime,
		ResponseTimeMean:         p.TotalResponseTime / p.TotalRequestCount,
		TotalFailureRequestCount: p.TotalFailureRequestCount,
		TotalRequestCount:        p.TotalRequestCount,
		ThreadNum:                p.ThreadNum,
	}
}

//func (p *PressStat) Log() {
//	p.Lock.Lock()
//	defer p.Lock.Unlock()
//	var (
//		qps              int64
//		meanQPS          int64
//		meanResponseTime int64
//	)
//	if p.Throughput > 0 {
//		qps = p.Throughput / time.Since(p.ThroughputLastCalculateTime).Milliseconds() * 1000
//	}
//	if p.TotalSuccessRequestCount > 0 {
//		meanQPS = p.TotalSuccessRequestCount / p.TotalStatCount
//	}
//	if p.TotalResponseTime > 0 {
//		meanResponseTime = p.TotalResponseTime / p.TotalStatCount
//	}
//	logrus.WithFields(logrus.Fields{
//		"QPS":      qps,
//		"QPS(平均)":  meanQPS,
//		"响应时间(最小)": p.MinResponseTime,
//		"响应时间(平均)": meanResponseTime,
//		"响应时间(最大)": p.MaxResponseTime,
//		"总失败数":     p.TotalFailureRequestCount,
//		"总请求数":     p.TotalRequestCount,
//		"线程数":      p.ThreadNum,
//	}).Info("pressing...")
//	p.Throughput = 0
//	p.ThroughputLastCalculateTime = time.Now()
//}

func (p *PressStat) Close() error {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	p.Closed = true
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
		p.ThroughputLastCalculateTime = time.Now()
	})
	if responseTime < p.MinResponseTime {
		p.MinResponseTime = responseTime
	}
	if responseTime > p.MaxResponseTime {
		p.MaxResponseTime = responseTime
	}
	p.TotalResponseTime += responseTime
}

func (p *PressStat) RecordThroughput(startTime time.Time) {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	if p.Closed {
		return
	}
	p.Throughput = int64(float64(p.Throughput) / float64(time.Since(p.ThroughputLastCalculateTime).Milliseconds()) * 1000)
}
