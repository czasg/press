package cmd

import (
    "context"
    "errors"
    "github.com/czasg/press"
    "github.com/spf13/cobra"
    "log"
    "os"
    "os/signal"
    "strings"
    "syscall"
)

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

func InitCobraCmd() *cobra.Command {
    cmd := NewPressCommand()
    InstallStartCommand(cmd)
    InstallTestCommand(cmd)
    InstallInitCommand(cmd)
    return cmd
}

func NewPressCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:  "press",
        Long: "press test tool",
        CompletionOptions: cobra.CompletionOptions{
            DisableDefaultCmd: true,
        },
    }
    return cmd
}

func InstallStartCommand(cmd *cobra.Command) {
    var file string
    startCmd := &cobra.Command{
        Use:   "start",
        Short: "start a press test",
        Long:  `start a press test with template file`,
        RunE: func(cmd *cobra.Command, args []string) error {
            var cfg press.Config
            if file == "" {
                return errors.New("file不能为空")
            }
            err := press.ParseConfig(file, &cfg)
            if err != nil {
                return err
            }
            ctx := NewSignalContext()
            press.RunPressCMD(ctx, &cfg)
            return nil
        },
    }
    cf := startCmd.Flags()
    cf.StringVarP(&file, "file", "f", "", "a press test template file")
    cmd.AddCommand(startCmd)
}

func InstallTestCommand(cmd *cobra.Command) {
    step := press.Steps{Name: "press test"}
    testCmd := &cobra.Command{
        Use:   "test",
        Short: "a short press test",
        Long:  `start a short press test with flags`,
        RunE: func(cmd *cobra.Command, args []string) error {
            if step.Http.Url == "" {
                return errors.New("url不能为空")
            }
            headers, err := cmd.Flags().GetStringArray("header")
            if err != nil {
                return err
            }
            step.Http.Headers = map[string]string{}
            for _, header := range headers {
                hs := strings.Split(header, ":")
                if len(hs) != 2 {
                    continue
                }
                step.Http.Headers[hs[0]] = hs[1]
            }
            cfg := press.Config{
                Version: "1",
                Metadata: press.Metadata{
                    Name: "press",
                },
            }
            cfg.Steps = append(cfg.Steps, step)
            ctx := NewSignalContext()
            press.RunPressCMD(ctx, &cfg)
            return nil
        },
    }
    cf := testCmd.Flags()
    {
        cf.IntVar(&step.LogInterval, "interval", 1, "log output interval")
    }
    {
        cf.IntVar(&step.ThreadGroup.Thread, "thread", 1, "press threads")
        cf.IntVar(&step.ThreadGroup.Duration, "duration", 10, "press duration")
        cf.IntVar(&step.ThreadGroup.ThreadRampUp, "ramp", 1, "press threads ramp-up")
    }
    {
        cf.StringVar(&step.Http.Url, "url", "", "http requests url")
        cf.StringVar(&step.Http.Method, "method", "GET", "http requests method, like GET,POST")
        cf.IntVar(&step.Http.Timeout, "timeout", 1, "http requests timeout")
        cf.StringVar(&step.Http.Body, "body", "", "http requests body")
        cf.StringArray("header", []string{}, "http requests headers, like content-type:application/json")
    }
    {
        cf.IntVar(&step.Assert.StatusCode, "statuscode", 200, "assert http response statuscode")
    }
    cmd.AddCommand(testCmd)
}

func InstallInitCommand(cmd *cobra.Command) {
    var file string
    initCmd := &cobra.Command{
        Use:   "init",
        Short: "initialize template file",
        Long:  `create a template if not exists`,
        RunE: func(cmd *cobra.Command, args []string) error {
            return press.CreateYaml(file)
        },
    }
    cf := initCmd.Flags()
    cf.StringVarP(&file, "file", "f", "press.yaml", "init a press test template file")
    cmd.AddCommand(initCmd)
}
