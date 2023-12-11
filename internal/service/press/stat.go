package press

import (
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type IStat interface {
	RecordSuccess()
	RecordFailure()
	RecordThread()
	RecordTime(startTime time.Time)
	RecordThroughput(startTime time.Time)
	Info() StatInfo
}

type StatInfo struct{}

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
}

func (p *PressStat) RecordSuccess() {
	p.Lock.Lock()
	p.TotalRequestCount++
	p.TotalSuccessRequestCount++
	p.Throughput++
	p.Lock.Unlock()
}

func (p *PressStat) RecordFailure() {
	p.Lock.Lock()
	p.TotalRequestCount++
	p.TotalFailureRequestCount++
	p.Lock.Unlock()
}

func (p *PressStat) RecordThread() {
	p.Lock.Lock()
	p.ThreadNum++
	p.Lock.Unlock()
}

func (p *PressStat) RecordTime(startTime time.Time) {
	p.Lock.Lock()
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
	p.Lock.Unlock()
}

func (p *PressStat) RecordThroughput(startTime time.Time) {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	p.Throughput = int64(float64(p.Throughput) / float64(time.Since(p.ThroughputLastCalculateTime).Milliseconds()) * 1000)
}

func (p *PressStat) Log() {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	var (
		qps              int64
		meanQPS          int64
		meanResponseTime int64
	)
	if p.Throughput > 0 {
		qps = p.Throughput / time.Since(p.ThroughputLastCalculateTime).Milliseconds() * 1000
	}
	if p.TotalSuccessRequestCount > 0 {
		meanQPS = p.TotalSuccessRequestCount / p.TotalStatCount
	}
	if p.TotalResponseTime > 0 {
		meanResponseTime = p.TotalResponseTime / p.TotalStatCount
	}
	logrus.WithFields(logrus.Fields{
		"QPS":      qps,
		"QPS(平均)":  meanQPS,
		"响应时间(最小)": p.MinResponseTime,
		"响应时间(平均)": meanResponseTime,
		"响应时间(最大)": p.MaxResponseTime,
		"总失败数":     p.TotalFailureRequestCount,
		"总请求数":     p.TotalRequestCount,
		"线程数":      p.ThreadNum,
	}).Info("pressing...")
	p.Throughput = 0
	p.ThroughputLastCalculateTime = time.Now()
}
