package cmd

import (
    "context"
    "errors"
    "fmt"
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
        Long: "压力测试工具",
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
        Short: "启动压力测试",
        Long:  `基于模板创建压力测试`,
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
    cf.StringVarP(&file, "file", "f", "", "压力测试模板文件")
    cmd.AddCommand(startCmd)
}

func InstallTestCommand(cmd *cobra.Command) {
    step := press.Steps{Name: "press test"}
    testCmd := &cobra.Command{
        Use:   "test [url]",
        Short: "快速测试",
        Long:  `基于指令的快速压力测试`,
        RunE: func(cmd *cobra.Command, args []string) error {
            if len(args) < 1 {
                return errors.New("url不能为空")
            }
            url := args[0]
            if !strings.HasPrefix(url, "http") {
                url = fmt.Sprintf("http://%s", url)
            }
            step.Http.Url = url
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
        cf.IntVar(&step.LogInterval, "interval", 1, "日志间隔")
    }
    {
        cf.IntVar(&step.ThreadGroup.Thread, "thread", 1, "线程数")
        cf.IntVar(&step.ThreadGroup.Duration, "duration", 10, "压力测试持续时间")
        cf.IntVar(&step.ThreadGroup.ThreadRampUp, "ramp", 1, "线程唤醒时间")
    }
    {
        cf.StringVar(&step.Http.Method, "method", "GET", "HTTP 请求方式")
        cf.IntVar(&step.Http.Timeout, "timeout", 1, "HTTP 请求超时时间")
        cf.StringVar(&step.Http.Body, "body", "", "HTTP 请求体")
        cf.StringArray("header", []string{}, "HTTP 请求头")
    }
    {
        cf.IntVar(&step.Assert.StatusCode, "statuscode", 200, "断言状态码")
    }
    cmd.AddCommand(testCmd)
}

func InstallInitCommand(cmd *cobra.Command) {
    var file string
    initCmd := &cobra.Command{
        Use:   "init",
        Short: "初始化模板文件",
        Long:  `快速初始化一个压力测试模板文件`,
        RunE: func(cmd *cobra.Command, args []string) error {
            return press.CreateYaml(file)
        },
    }
    cf := initCmd.Flags()
    cf.StringVarP(&file, "file", "f", "press.yaml", "压力测试模板")
    cmd.AddCommand(initCmd)
}
