package subcmd

import (
	"context"
	"fmt"
	"github.com/czasg/press/internal/config"
	"github.com/czasg/press/internal/service"
	"github.com/czasg/press/internal/utils"
	"github.com/spf13/cobra"
	"os"
)

func NewPressStartCommand(ctx context.Context) *cobra.Command {
	//var file string
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "start a press test by config yaml file",
		Long:  `start a press test by config yaml file`,
		RunE: func(cmd *cobra.Command, args []string) error {
			file, err := cmd.Flags().GetString("file")
			if err != nil {
				return err
			}
			if !utils.FileExist(file) {
				return fmt.Errorf("file[%s] not found", file)
			}
			body, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			cfg, err := config.Parse(body)
			if err != nil {
				return err
			}
			return service.RunPress(ctx, cfg)
		},
	}
	cf := startCmd.Flags()
	{
		cf.StringP("file", "f", "press.yaml", "压力测试模板文件")
	}
	startCmd.AddCommand(NewPressStartManagerCommand(ctx))
	startCmd.AddCommand(NewPressStartWorkerCommand(ctx))
	return startCmd
}

func NewPressStartManagerCommand(ctx context.Context) *cobra.Command {
	ManagerCmd := &cobra.Command{
		Use:   "manager",
		Short: "start a press manager",
		Long:  `start a press manager`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return ManagerCmd
}

func NewPressStartWorkerCommand(ctx context.Context) *cobra.Command {
	workerCmd := &cobra.Command{
		Use:   "worker",
		Short: "start a press worker",
		Long:  `start a press worker`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return workerCmd
}
