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
    Count             int64
    OkQPS             int64
    Ok                int64
    Kill              int64
    MinResponseTime   int64
    MaxResponseTime   int64
    TotalResponseTime int64
}

func (this *Stat) RecordOK() {
    this.lock.Lock()
    this.TotalRequests++
    this.OkQPS++
    this.Ok++
    this.lock.Unlock()
}

func (this *Stat) RecordKill() {
    this.lock.Lock()
    this.TotalRequests++
    this.Kill++
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
    this.Count++
    record := Record{
        RecordTime:      time.Now(),
        TotalRequests:   this.TotalRequests,
        CurrentQPS:      this.OkQPS,
        MeanQPS:         this.Ok / this.Count,
        KillRequests:    this.Kill,
        MinResponseTime: this.MinResponseTime,
        MaxResponseTime: this.MaxResponseTime,
    }
    if this.TotalRequests > 0 {
        record.MeanResponseTime = this.TotalResponseTime / this.TotalRequests
    }
    this.OkQPS = 0
    this.lock.Unlock()
    this.Save(record)
    return record.String()
}

func (this *Stat) Save(record Record) {
}

type Record struct {
    RecordTime       time.Time
    TotalRequests    int64
    CurrentQPS       int64
    MeanQPS          int64
    KillRequests     int64
    MinResponseTime  int64
    MaxResponseTime  int64
    MeanResponseTime int64
}

func (this Record) String() string {
    return fmt.Sprintf(
        "瞬时[%v]QPS 平均[%v]QPS 平均响应[%v]ms 最小/大响应[%v/%v]ms  总请求次数[%v] 错误次数[%v]",
        this.CurrentQPS,
        this.MeanQPS,
        this.MeanResponseTime,
        this.MinResponseTime,
        this.MaxResponseTime,
        this.TotalRequests,
        this.KillRequests,
    )
}
