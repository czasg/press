package subcmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/czasg/press/internal/service/press"
	"github.com/czasg/press/internal/yamltemplate"
	"github.com/spf13/cobra"
	"strings"
)

func NewPressTestCommand(ctx context.Context) *cobra.Command {
	step := yamltemplate.StepsV1{Name: "press test"}
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
			fmt.Printf("%#v\n", step.Http)
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
			return press.RunPressV1(ctx, yamltemplate.ConfigV1{
				Version: "1",
				Metadata: yamltemplate.MetadataV1{
					Name: "press test",
				},
				Steps: []yamltemplate.StepsV1{
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
		cf.IntVar(&step.Http.Timeout, "timeout", 24, "HTTP 请求超时时间")
		cf.StringVar(&step.Http.Body, "body", "", "HTTP 请求体")
		cf.StringArray("header", []string{}, "HTTP 请求头")
	}
	{
		cf.IntVar(&step.Assert.StatusCode, "statuscode", 200, "断言状态码")
	}
	return testCmd
}
