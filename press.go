package press

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "github.com/sirupsen/logrus"
    "io/ioutil"
    "log"
    "net"
    "net/http"
    "strings"
    "time"
)

func RunPressCMD(ctx context.Context, cfg *Config) {
    logrus.WithField("Version", cfg.Version).Info("检测到当前版本")
    logrus.WithField("User", cfg.Metadata.Name).Info("检测到当前用户")
    for index, step := range cfg.Steps {
        logrus.Info("#########################")
        logrus.Printf("###### 任务[%v]开始 ######", index)
        logrus.Info("#########################")
        logrus.Printf("名称：%v", step.Name)
        logrus.Printf("线程数：%v", step.ThreadGroup.Thread)
        logrus.Printf("线程唤醒时间：%vs", step.ThreadGroup.ThreadRampUp)
        logrus.Printf("持续时间：%vs", step.ThreadGroup.Duration)
        logrus.Printf("日志输出间隔：%vs", step.LogInterval)
        runPressHttp(ctx, step)
    }
    logrus.Info("##########################")
    logrus.Info("###### 压力测试结束 ######")
    logrus.Info("##########################")
}

func runPressHttp(ctx context.Context, step Steps) {
    ctxTime, _ := context.WithTimeout(ctx, time.Duration(step.ThreadGroup.Duration)*time.Second)
    stat := NewStat(step)
    defer func() {
        if stat.OutputFile != nil {
            stat.OutputFile.Close()
        }
    }()
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
