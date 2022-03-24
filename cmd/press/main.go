package main

import (
    "context"
    "flag"
    "fmt"
    "github.com/czasg/press"
    "log"
    "os"
    "os/signal"
    "syscall"
)

func CreateYaml() {
    f, err := os.Create("press.yaml")
    if err != nil {
        log.Printf("初始化yaml文件异常：%v", err)
        return
    }
    defer f.Close()
    _, err = f.WriteString(press.TemplateV1)
    if err != nil {
        log.Printf("初始化yaml内容异常：%v", err)
        return
    }
    log.Printf("初始化 yaml 成功")
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
    flag.Usage = func() {
        fmt.Println(`压力测试工具 press 使用方法：
* 初始化 press.yaml 文件：
    press init
* 测试开始：
    press -f press.yaml
* 参数说明：`)
        flag.PrintDefaults()
    }
    cfgFilePath := flag.String("f", "", "配置文件")
    flag.Parse()
    if flag.Arg(0) == "init" {
        CreateYaml()
        return
    }
    if *cfgFilePath == "" {
        flag.PrintDefaults()
        log.Fatalf("未指定配置文件")
    }
    ctx := NewSignalContext()
    cfg := press.MustParseConfig(*cfgFilePath)
    press.RunPressCMD(ctx, cfg)
}
