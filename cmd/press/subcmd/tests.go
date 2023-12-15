package subcmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/czasg/press/internal/config"
	"github.com/czasg/press/internal/service"
	"github.com/czasg/snow"
	"github.com/spf13/cobra"
	"strings"
)

func NewPressTestCommand(ctx context.Context) *cobra.Command {
	step := config.Steps{}
	testCmd := &cobra.Command{
		Use:   "test [url]",
		Short: "press testing",
		Long: `press testing, eg:
- press test localhost:8080
- press test localhost:8080 --method=POST --header=Content-Type:application/json
- press test -thread=2 --duration=30
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("Empty Url")
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
				hs := strings.Split(strings.TrimSpace(header), ":")
				if len(hs) != 2 {
					continue
				}
				step.Http.Headers[strings.Title(hs[0])] = strings.TrimSpace(hs[1])
			}
			return service.RunPress(ctx, &config.Config{
				Version: "1",
				Metadata: config.Metadata{
					Name: "press test",
					Uid:  snow.Next(),
				},
				Steps: []config.Steps{
					step,
				},
			})
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
		cf.Float64Var(&step.Http.Timeout, "timeout", 24, "HTTP 请求超时时间")
		cf.BoolVar(&step.Http.Keepalive, "keepalive", false, "HTTP Keepalive")
		cf.StringVar(&step.Http.Body, "body", "", "HTTP 请求体")
		cf.StringArray("header", []string{}, "HTTP 请求头")
	}
	{
		cf.IntVar(&step.Assert.StatusCode, "statuscode", 200, "断言状态码")
	}
	return testCmd
}
