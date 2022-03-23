package basis

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

func (s *Stat) RecordOK() {
	s.lock.Lock()
	s.OkQPS++
	s.Ok++
	s.lock.Unlock()
}

func (s *Stat) RecordKill() {
	s.lock.Lock()
	s.KillQPS++
	s.Kill++
	s.lock.Unlock()
}

func (s *Stat) RecordResponseTime(start time.Time) {
	s.lock.Lock()
	s.ResponseTime += time.Since(start).Milliseconds()
	s.lock.Unlock()
}

func (s *Stat) String() string {
	s.lock.Lock()
	s.Count++
	text := fmt.Sprintf(
		"瞬时：[%v]QPS 平均：[%v]QPS 平均响应：[%v]ms 错误次数：[%v]",
		s.OkQPS,
		s.Ok/s.Count,
		s.ResponseTime/s.Count,
		s.Kill,
	)
	s.OkQPS = 0
	s.KillQPS = 0
	s.lock.Unlock()
	return text
}

func (s *Stat) Save(output Output) string {
	s.lock.Lock()
	if output.Path == "" {
		return "未指定存储路径"
	}
	s.lock.Unlock()
	return "暂不支持"
}
