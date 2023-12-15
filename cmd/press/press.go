package main

import (
	"context"
	"github.com/czasg/press/cmd/press/subcmd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx := NewSignalContext()
	cmd := NewPressCommand()
	cmd.AddCommand(subcmd.NewPressInitCommand(ctx))  // press init
	cmd.AddCommand(subcmd.NewPressStartCommand(ctx)) // press start
	cmd.AddCommand(subcmd.NewPressTestCommand(ctx))  // press test
	err := cmd.Execute()
	if err != nil {
		logrus.Panic(err)
	}
}

func NewSignalContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
		logrus.Warnf("detected system exit signal: [%v]", <-ch)
		cancel()
	}()
	return ctx
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

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
}
