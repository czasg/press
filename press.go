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
        func() {
            ctxTime, _ := context.WithTimeout(ctx, time.Duration(step.ThreadGroup.Duration)*time.Second)
            output := &Output{}
            //wg := sync.WaitGroup{}
            press := func() {
                //defer wg.Done()
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
                    resp, _ := client.Do(req)
                    body, _ := ioutil.ReadAll(resp.Body)
                    _ = resp.Body.Close()
                    if step.Assert.StatusCode > 0 && resp.StatusCode != step.Assert.StatusCode {
                        output.RecordKill()
                        continue
                    }
                    if len(step.Assert.Headers) > 0 {
                        for _, header := range step.Assert.Headers {
                            for k, v := range header {
                                if resp.Header.Get(k) != v {
                                    output.RecordKill()
                                    continue
                                }
                            }
                        }
                    }
                    if step.Assert.Body != "" && string(body) != step.Assert.Body {
                        output.RecordKill()
                        continue
                    }
                    if len(step.Assert.JsonMap) > 0 {
                        var m map[string]interface{}
                        err := json.Unmarshal(body, &m)
                        if err != nil {
                            output.RecordKill()
                            continue
                        }
                        for _, jsonMap := range step.Assert.JsonMap {
                            for k, v := range jsonMap {
                                v1, ok := m[k]
                                if !ok {
                                    output.RecordKill()
                                    continue
                                }
                                if !reflect.DeepEqual(v, v1) {
                                    output.RecordKill()
                                    continue
                                }
                            }
                        }
                    }
                    output.RecordOK()
                }
            }

            for i := 0; i < step.ThreadGroup.Thread; i++ {
                //wg.Add(1)
                go press()
            }
            for {
                select {
                case <-ctxTime.Done():
                    return
                default:
                    time.Sleep(time.Second)
                    log.Println(output.String())
                }
            }
        }()

        //ctxTime, _ := context.WithTimeout(ctx, time.Duration(step.ThreadGroup.Duration)*time.Second)
        //output := &Output{}
        //wg := sync.WaitGroup{}
        //press := func() {
        //    defer wg.Done()
        //    client := &http.Client{
        //        Transport: &http.Transport{
        //            DialContext: (&net.Dialer{
        //                KeepAlive: time.Hour * 24,
        //            }).DialContext,
        //        },
        //    }
        //    for {
        //        select {
        //        case <-ctxTime.Done():
        //            return
        //        default:
        //        }
        //        req, _ := http.NewRequest(step.Http.Method, step.Http.Url, bytes.NewBuffer([]byte(step.Http.Body)))
        //        for k, v := range step.Http.Headers {
        //            req.Header.Add(k, v)
        //        }
        //        resp, _ := client.Do(req)
        //        body, _ := ioutil.ReadAll(resp.Body)
        //        _ = resp.Body.Close()
        //        if step.Assert.StatusCode > 0 && resp.StatusCode != step.Assert.StatusCode {
        //            output.RecordKill()
        //            continue
        //        }
        //        if len(step.Assert.Headers) > 0 {
        //            for _, header := range step.Assert.Headers {
        //                for k, v := range header {
        //                    if resp.Header.Get(k) != v {
        //                        output.RecordKill()
        //                        continue
        //                    }
        //                }
        //            }
        //        }
        //        if step.Assert.Body != "" && string(body) != step.Assert.Body {
        //            output.RecordKill()
        //            continue
        //        }
        //        if len(step.Assert.JsonMap) > 0 {
        //            var m map[string]interface{}
        //            err := json.Unmarshal(body, &m)
        //            if err != nil {
        //                output.RecordKill()
        //                continue
        //            }
        //            for _, jsonMap := range step.Assert.JsonMap {
        //                for k, v := range jsonMap {
        //                    v1, ok := m[k]
        //                    if !ok {
        //                        output.RecordKill()
        //                        continue
        //                    }
        //                    if !reflect.DeepEqual(v, v1) {
        //                        output.RecordKill()
        //                        continue
        //                    }
        //                }
        //            }
        //        }
        //        output.RecordOK()
        //    }
        //}
        //
        //for i := 0; i < step.ThreadGroup.Thread; i++ {
        //    wg.Add(1)
        //    go press()
        //}
        //for {
        //    select {
        //    case <-ctxTime.Done():
        //        return
        //    default:
        //        time.Sleep(time.Second)
        //        log.Println(output.String())
        //    }
        //}
        //wg.Wait()
    }
}
