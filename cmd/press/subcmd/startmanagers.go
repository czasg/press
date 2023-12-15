package subcmd

import (
	"context"
	"github.com/czasg/press/internal/config"
	"github.com/czasg/press/third"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func NewPressStartManagerCommand(ctx context.Context) *cobra.Command {
	managerCmd := &cobra.Command{
		Use:   "manager",
		Short: "start a press manager",
		Long:  `start a press manager`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := readConfig(cmd)
			if err != nil {
				return err
			}
			rds, err := third.NewRedis(cfg)
			if err != nil {
				return err
			}
			channelName := cfg.Metadata.Annotations.PressClusterBrokerRedisPbWorker
			for _, step := range cfg.Steps {
				newCfg, err := readConfig(cmd)
				if err != nil {
					return err
				}
				newCfg.Steps = []config.Steps{
					step,
				}
				body, err := yaml.Marshal(newCfg)
				if err != nil {
					return err
				}
				rds.WithContext(ctx).Publish(channelName, body)
			}
			return nil
		},
	}
	cf := managerCmd.Flags()
	{
		cf.StringP("file", "f", "press.yaml", "压力测试配置文件")
	}
	return managerCmd
}
