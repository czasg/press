package subcmd

import (
	"context"
	"github.com/spf13/cobra"
)

func NewPressStartManagerCommand(ctx context.Context) *cobra.Command {
	managerCmd := &cobra.Command{
		Use:   "manager",
		Short: "start a press manager",
		Long:  `start a press manager`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := readConfig(cmd)
			if err != nil {
				return err
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
