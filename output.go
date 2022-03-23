package press

import (
    "fmt"
    "sync"
    "time"
)

type Stat struct {
    lock         sync.Mutex
    Count        int64
    OkQPS        int64
    KillQPS      int64
    Ok           int64
    Kill         int64
    ResponseTime int64
}

func (this *Stat) RecordOK() {
    this.lock.Lock()
    this.OkQPS++
    this.Ok++
    this.lock.Unlock()
}

func (this *Stat) RecordKill() {
    this.lock.Lock()
    this.KillQPS++
    this.Kill++
    this.lock.Unlock()
}

func (this *Stat) RecordResponseTime(start time.Time) {
    this.lock.Lock()
    this.ResponseTime += time.Since(start).Milliseconds()
    this.lock.Unlock()
}

func (this *Stat) String() string {
    this.lock.Lock()
    this.Count++
    text := fmt.Sprintf(
        "瞬时：[%v]QPS 平均：[%v]QPS 平均响应：[%v] 错误次数：[%v]",
        this.OkQPS,
        this.Ok/this.Count,
        this.ResponseTime/this.Count,
        this.Kill,
    )
    this.OkQPS = 0
    this.KillQPS = 0
    this.lock.Unlock()
    return text
}

func (this *Stat) Save(output Output) string {
    this.lock.Lock()
    if output.Path == "" {
        return "未指定存储路径"
    }
    this.lock.Unlock()
    return "暂不支持"
}
