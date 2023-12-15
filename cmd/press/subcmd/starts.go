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
			cfg, err := readConfig(cmd)
			if err != nil {
				return err
			}
			return service.RunPress(ctx, cfg)
		},
	}
	cf := startCmd.Flags()
	{
		cf.StringP("file", "f", "press.yaml", "压力测试配置文件")
	}
	startCmd.AddCommand(NewPressStartManagerCommand(ctx))
	startCmd.AddCommand(NewPressStartWorkerCommand(ctx))
	return startCmd
}

func readConfig(cmd *cobra.Command) (*config.Config, error) {
	file, err := cmd.Flags().GetString("file")
	if err != nil {
		return nil, err
	}
	if !utils.FileExist(file) {
		return nil, fmt.Errorf("file[%s] not found", file)
	}
	body, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	cfg, err := config.Parse(body)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
