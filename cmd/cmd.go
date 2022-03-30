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
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "start a press test",
		Long:  `start a press test with template`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var cfg press.Config
			ctx := NewSignalContext()
			file, err := cmd.Flags().GetString("file")
			if err != nil {
				return err
			}
			if file == "" {
				return errors.New("file不能为空")
			}
			err = press.ParseConfig(file, &cfg)
			if err != nil {
				return err
			}
			press.RunPressCMD(ctx, &cfg)
			return nil
		},
	}
	cf := startCmd.Flags()
	cf.StringP("file", "f", "", "press template file")
	cmd.AddCommand(startCmd)
}

func InstallTestCommand(cmd *cobra.Command) {
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "test a short press test",
		Long:  `start a short press test with flags`,
		RunE: func(cmd *cobra.Command, args []string) error {
			url, err := cmd.Flags().GetString("url")
			if err != nil {
				return err
			}
			if url == "" {
				return errors.New("url不能为空")
			}
			method, err := cmd.Flags().GetString("method")
			if err != nil {
				return err
			}
			headers, err := cmd.Flags().GetStringArray("header")
			if err != nil {
				return err
			}
			headersMap := map[string]string{}
			for _, header := range headers {
				hs := strings.Split(header, ":")
				if len(hs) != 2 {
					continue
				}
				headersMap[hs[0]] = hs[1]
			}
			body, err := cmd.Flags().GetString("body")
			if err != nil {
				return err
			}
			thread, err := cmd.Flags().GetInt("thread")
			if err != nil {
				return err
			}
			duration, err := cmd.Flags().GetInt("duration")
			if err != nil {
				return err
			}
			step := press.Steps{
				Name:        "test",
				LogInterval: 1,
				ThreadGroup: press.ThreadGroup{
					Thread:       thread,
					ThreadRampUp: 1,
					Duration:     duration,
				},
				Http: press.Http{
					Url:     url,
					Method:  strings.ToUpper(method),
					Timeout: 60,
					Headers: headersMap,
					Body:    body,
				},
				Assert: press.Assert{
					StatusCode: 200,
				},
				Output: press.Output{},
			}
			var cfg press.Config
			cfg.Steps = append(cfg.Steps, step)
			ctx := NewSignalContext()
			press.RunPressCMD(ctx, &cfg)
			return nil
		},
	}
	cf := testCmd.Flags()
	cf.String("url", "", "requests url")
	cf.String("method", "GET", "requests method, like GET,POST")
	cf.String("body", "", "requests body")
	cf.StringArray("header", []string{}, "requests header")
	cf.Int("thread", 1, "requests threads")
	cf.Int("duration", 10, "requests duration")
	cmd.AddCommand(testCmd)
}

func InstallInitCommand(cmd *cobra.Command) {
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "initialize template",
		Long:  `create a template if not exists`,
		RunE: func(cmd *cobra.Command, args []string) error {
			file, err := cmd.Flags().GetString("file")
			if err != nil {
				return err
			}
			return press.CreateYaml(file)
		},
	}
	cf := initCmd.Flags()
	cf.StringP("file", "f", "press.yaml", "point a template file")
	cmd.AddCommand(initCmd)
}
