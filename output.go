package press

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func NewStat(step Steps) *Stat {
	stat := &Stat{}
	if step.Output.Path != "" {
		result := filepath.Join(step.Output.Path, fmt.Sprintf("%v.press", step.Name))
		f, err := createFile(result)
		if err != nil {
			log.Fatalln(err)
		}
		stat.OutputFile = f
	}
	return stat
}

func createFile(filename string) (*os.File, error) {
	_, err := os.Stat(filename)
	if err == nil || os.IsExist(err) {
		return createFile(fmt.Sprintf("%v.%v", filename, time.Now().Unix()))
	}
	f, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("创建文件[%s]异常: %v", filename, err)
	}
	return f, nil
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
	OutputFile        *os.File
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
	if this.OutputFile == nil {
		return
	}
	body, _ := json.Marshal(record)
	_, _ = this.OutputFile.Write(body)
	_, _ = this.OutputFile.WriteString("\n")
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
		"瞬时/平均[%v/%v]QPS 最小/平均/最大响应[%v/%v/%v]ms 错误/总请求次数[%v/%v] 线程数[%v]",
		this.Throughput,
		this.MeanThroughput,
		this.MinResponseTime,
		this.MeanResponseTime,
		this.MaxResponseTime,
		this.ErrorRequests,
		this.TotalRequests,
		this.Threads,
	)
}
