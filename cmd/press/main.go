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

func NewSignalContext() context.Context {
    ctx, cancel := context.WithCancel(context.Background())
    go func() {
        ch := make(chan os.Signal, 1)
        signal.Notify(ch, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
        <-ch
        cancel()
    }()
    return ctx
}

func main() {
    cfgFilePath := flag.String("f", "", "配置文件")
    flag.Parse()
    if *cfgFilePath == "" {
        flag.PrintDefaults()
        log.Fatalf("未指定配置文件")
    }
    cfg := press.MustParseConfig(*cfgFilePath)
    press.RunPressV1(NewSignalContext(), cfg)
}
