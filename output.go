package press

import (
    "fmt"
    "sync"
    "time"
)

func NewStat(step Steps) *Stat {
    return &Stat{}
}

type Stat struct {
    lock              sync.Mutex
    once              sync.Once
    TotalRequests     int64
    StatCount         int64
    Throughput        int64
    SuccessRequests   int64
    ErrorRequests     int64
    MinResponseTime   int64
    MaxResponseTime   int64
    TotalResponseTime int64
    Threads           int64
}

func (this *Stat) RecordSuccess() {
    this.lock.Lock()
    this.TotalRequests++
    this.Throughput++
    this.SuccessRequests++
    this.lock.Unlock()
}

func (this *Stat) RecordError() {
    this.lock.Lock()
    this.TotalRequests++
    this.ErrorRequests++
    this.lock.Unlock()
}

func (this *Stat) RecordThread() {
    this.lock.Lock()
    this.Threads++
    this.lock.Unlock()
}

func (this *Stat) RecordResponseTime(start time.Time) {
    this.lock.Lock()
    responseTime := time.Since(start).Milliseconds()
    this.once.Do(func() {
        if this.MinResponseTime == 0 {
            this.MinResponseTime = responseTime
        }
    })
    if responseTime < this.MinResponseTime {
        this.MinResponseTime = responseTime
    }
    if responseTime > this.MaxResponseTime {
        this.MaxResponseTime = responseTime
    }
    this.TotalResponseTime += responseTime
    this.lock.Unlock()
}

func (this *Stat) String() string {
    this.lock.Lock()
    this.StatCount++
    record := Record{
        RecordTime:      time.Now(),
        TotalRequests:   this.TotalRequests,
        Throughput:      this.Throughput,
        MeanThroughput:  this.SuccessRequests / this.StatCount,
        ErrorRequests:   this.ErrorRequests,
        MinResponseTime: this.MinResponseTime,
        MaxResponseTime: this.MaxResponseTime,
        Threads:         this.Threads,
    }
    if this.TotalRequests > 0 {
        record.MeanResponseTime = this.TotalResponseTime / this.TotalRequests
    }
    this.Throughput = 0
    this.lock.Unlock()
    this.Save(record)
    return record.String()
}

func (this *Stat) Save(record Record) {
}

type Record struct {
    RecordTime       time.Time
    TotalRequests    int64
    Throughput       int64
    MeanThroughput   int64
    ErrorRequests    int64
    MinResponseTime  int64
    MaxResponseTime  int64
    MeanResponseTime int64
    Threads          int64
}

func (this Record) String() string {
    return fmt.Sprintf(
        "瞬时[%v]QPS 平均[%v]QPS 平均响应[%v]ms 最小/大响应[%v/%v]ms 总请求次数[%v] 错误次数[%v] 线程数[%v]",
        this.Throughput,
        this.MeanThroughput,
        this.MeanResponseTime,
        this.MinResponseTime,
        this.MaxResponseTime,
        this.TotalRequests,
        this.ErrorRequests,
        this.Threads,
    )
}
