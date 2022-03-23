package press

import (
    "bytes"
    "context"
    "encoding/json"
    "io/ioutil"
    "log"
    "net"
    "net/http"
    "reflect"
    "time"
)

func RunPressV1(ctx context.Context, cfg *Config) {
    log.Printf("当前版本：%v\n", cfg.Version)
    log.Printf("当前用户：%v\n", cfg.Metadata.Name)
    for index, step := range cfg.Steps {
        log.Printf("----- 任务[%v]开始 -----\n", index)
        log.Printf("名称：%v\n", step.Name)
        log.Printf("线程数：%v\n", step.ThreadGroup.Thread)
        log.Printf("线程唤醒时间：%vs\n", step.ThreadGroup.ThreadRampUp)
        log.Printf("持续时间：%vs\n", step.ThreadGroup.Duration)
        log.Printf("日志输出间隔：%vs\n", step.LogInterval)
        func() {
            ctxTime, _ := context.WithTimeout(ctx, time.Duration(step.ThreadGroup.Duration)*time.Second)
            stat := &Stat{}
            press := func() {
                client := &http.Client{
                    Transport: &http.Transport{
                        DialContext: (&net.Dialer{
                            KeepAlive: time.Hour * 24,
                        }).DialContext,
                    },
                }
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
                        stat.RecordKill()
                        continue
                    }
                    stat.RecordResponseTime(start)
                    body, err := ioutil.ReadAll(resp.Body)
                    if err != nil {
                        stat.RecordKill()
                        continue
                    }
                    _ = resp.Body.Close()
                    if step.Assert.StatusCode > 0 && resp.StatusCode != step.Assert.StatusCode {
                        stat.RecordKill()
                        continue
                    }
                    if len(step.Assert.Headers) > 0 {
                        for _, header := range step.Assert.Headers {
                            for k, v := range header {
                                if resp.Header.Get(k) != v {
                                    stat.RecordKill()
                                    continue
                                }
                            }
                        }
                    }
                    if step.Assert.Body != "" && string(body) != step.Assert.Body {
                        stat.RecordKill()
                        continue
                    }
                    if len(step.Assert.JsonMap) > 0 {
                        var m map[string]interface{}
                        err := json.Unmarshal(body, &m)
                        if err != nil {
                            stat.RecordKill()
                            continue
                        }
                        for _, jsonMap := range step.Assert.JsonMap {
                            for k, v := range jsonMap {
                                v1, ok := m[k]
                                if !ok {
                                    stat.RecordKill()
                                    continue
                                }
                                if !reflect.DeepEqual(v, v1) {
                                    stat.RecordKill()
                                    continue
                                }
                            }
                        }
                    }
                    stat.RecordOK()
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
            interval := time.NewTicker(time.Second * time.Duration(step.LogInterval))
            for {
                select {
                case <-ctxTime.Done():
                    log.Printf("保存压测结果...")
                    log.Printf("保存完成[%v]", stat.Save(step.Output))
                    return
                case <-interval.C:
                    log.Println(stat.String())
                default:
                    _ = stat.String()
                }
                time.Sleep(time.Second)
            }
        }()
    }
    log.Printf("----- 压测结束 -----\n")
}
