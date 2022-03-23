package main

import (
    "context"
    "flag"
    "github.com/czasg/press"
    "log"
    "os"
    "os/signal"
    "syscall"
)

func CreateYaml() {
    f, err := os.Create("press.yaml")
    if err != nil {
        return
    }
    defer f.Close()
    f.WriteString(`---
version: "1"
metadata:
  name: "press"
steps:
  - name: "压力测试"
    logInterval: 1
    threadGroup:
      thread: 2
      threadRampUp: 1
      duration: 10
    http:
      url: "http://www.baidu.com"
      method: "POST"
      headers:
        content-type: "application/json"
      body: |
        {
          "hello":"press"
        }
    assert:
      statusCode: 200
      headers:
        - cookie: ""
      body: ""
#      jsonMap:
#        - key1: 1
#        - key2: 2
#    output:
#      path: "."
`)
}

func NewSignalContext() context.Context {
    ctx, cancel := context.WithCancel(context.Background())
    go func() {
        ch := make(chan os.Signal, 1)
        signal.Notify(ch, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
        log.Printf("检测到退出指令[%v]", <-ch)
        cancel()
    }()
    return ctx
}

func main() {
    version := flag.String("init", "", "初始化配置文件")
    cfgFilePath := flag.String("f", "", "配置文件")
    flag.Parse()
    if *version != "" {
        switch *version {
        case "1":
            CreateYaml()
        default:
            CreateYaml()
        }
        return
    }
    if *cfgFilePath == "" {
        flag.PrintDefaults()
        log.Fatalf("未指定配置文件")
    }
    cfg := press.MustParseConfig(*cfgFilePath)
    press.RunPressV1(NewSignalContext(), cfg)
}
