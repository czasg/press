package subcmd

import (
	"context"
	"fmt"
	"github.com/czasg/press/internal/utils"
	"github.com/czasg/press/internal/yamltemplate"
	"github.com/spf13/cobra"
	"os"
)

func NewPressStartCommand(ctx context.Context) *cobra.Command {
	//var file string
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "启动压力测试",
		Long:  `基于模板创建压力测试`,
		RunE: func(cmd *cobra.Command, args []string) error {
			file, err := cmd.Flags().GetString("file")
			if err != nil {
				return err
			}
			if !utils.FileExist(file) {
				return fmt.Errorf("文件[%s]不存在", file)
			}
			body, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			_, err = yamltemplate.ParseConfigV1(body)
			if err != nil {
				return err
			}
			//err = press.ParseConfig(file, &cfg)
			//if err != nil {
			//	return err
			//}
			//press.RunPressCMD(ctx, &cfg)
			return nil
		},
	}
	cf := startCmd.Flags()
	{
		cf.StringP("file", "f", "press.yaml", "压力测试模板文件")
	}
	return startCmd
}
