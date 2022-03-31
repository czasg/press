package press

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net"
    "net/http"
    "strings"
    "time"
)

func RunPressCMD(ctx context.Context, cfg *Config) {
    log.Printf("当前版本：%v\n", cfg.Version)
    log.Printf("当前用户：%v\n", cfg.Metadata.Name)
    for index, step := range cfg.Steps {
        log.Printf("----- 任务[%v]开始 -----\n", index)
        log.Printf("名称：%v\n", step.Name)
        log.Printf("线程数：%v\n", step.ThreadGroup.Thread)
        log.Printf("线程唤醒时间：%vs\n", step.ThreadGroup.ThreadRampUp)
        log.Printf("持续时间：%vs\n", step.ThreadGroup.Duration)
        log.Printf("日志输出间隔：%vs\n", step.LogInterval)
        runPressHttp(ctx, step)
    }
    log.Printf("----- 压测结束 -----\n")
}

func runPressHttp(ctx context.Context, step Steps) {
    ctxTime, _ := context.WithTimeout(ctx, time.Duration(step.ThreadGroup.Duration)*time.Second)
    stat := NewStat(step)
    press := func() {
        stat.RecordThread()
        client := &http.Client{
            Transport: &http.Transport{
                DialContext: (&net.Dialer{
                    KeepAlive: time.Second * time.Duration(step.ThreadGroup.Duration+1),
                }).DialContext,
            },
            Timeout: time.Second * time.Duration(step.Http.Timeout),
        }
    LOOP:
        for {
            select {
            case <-ctxTime.Done():
                return
            default:
            }
            req, _ := http.NewRequest(step.Http.Method, step.Http.Url, bytes.NewBuffer([]byte(step.Http.Body)))
            for k, v := range step.Http.Headers {
                req.Header.Add(k, v)
            }
            start := time.Now()
            resp, err := client.Do(req)
            if err != nil {
                stat.RecordError()
                continue
            }
            stat.RecordResponseTime(start)
            body, err := ioutil.ReadAll(resp.Body)
            if err != nil {
                stat.RecordError()
                continue
            }
            _ = resp.Body.Close()
            if step.Assert.StatusCode > 0 && resp.StatusCode != step.Assert.StatusCode {
                stat.RecordError()
                continue
            }
            if len(step.Assert.Headers) > 0 {
                for _, header := range step.Assert.Headers {
                    for k, v := range header {
                        if resp.Header.Get(k) != v {
                            stat.RecordError()
                            goto LOOP
                        }
                    }
                }
            }
            if step.Assert.Body != "" && string(body) != step.Assert.Body {
                stat.RecordError()
                continue
            }
            if len(step.Assert.JsonMap) > 0 {
                var m map[string]interface{}
                err := json.Unmarshal(body, &m)
                if err != nil {
                    stat.RecordError()
                    continue
                }
                for _, jsonMap := range step.Assert.JsonMap {
                    for k, v := range jsonMap {
                        v1, ok := m[k]
                        if !ok {
                            stat.RecordError()
                            goto LOOP
                        }
                        if !strings.EqualFold(
                            fmt.Sprintf("%v", v),
                            fmt.Sprintf("%v", v1),
                        ) {
                            stat.RecordError()
                            goto LOOP
                        }
                    }
                }
            }
            stat.RecordSuccess()
        }
    }
    go func() {
        interval := time.Second * time.Duration(step.ThreadGroup.ThreadRampUp) / time.Duration(step.ThreadGroup.Thread)
        for i := 0; i < step.ThreadGroup.Thread; i++ {
            select {
            case <-ctxTime.Done():
                return
            default:
                go press()
                time.Sleep(interval)
            }
        }
    }()
    var statLog = stat.String()
    second := time.NewTicker(time.Second)
    interval := time.NewTicker(time.Second * time.Duration(step.LogInterval))
    for {
        select {
        case <-ctxTime.Done():
            return
        case <-second.C:
            statLog = stat.String()
        case <-interval.C:
            log.Println(statLog)
        }
    }
}
