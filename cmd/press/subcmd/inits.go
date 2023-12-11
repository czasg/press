package subcmd

import (
	"context"
	"fmt"
	"github.com/czasg/press/internal/utils"
	"github.com/czasg/press/internal/yamltemplate"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

type PressInitCommandFlag struct {
	File        string `json:"file" yaml:"file"`
	Version     string `json:"version" yaml:"version"`
	AutoConfirm bool   `json:"autoConfirm" yaml:"autoConfirm"`
}

func NewPressInitCommand(ctx context.Context) *cobra.Command {
	flag := PressInitCommandFlag{}
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "初始化模板文件",
		Long:  `快速初始化一个压力测试模板文件`,
		RunE: func(cmd *cobra.Command, args []string) error {
			logs := logrus.WithFields(logrus.Fields{
				"version": flag.Version,
				"file":    flag.File,
			})
			if !filepath.IsAbs(flag.File) {
				filename, err := filepath.Abs(flag.File)
				if err != nil {
					return err
				}
				logs.Infof("修正文件绝对路径[%s]->[%s]", flag.File, filename)
				flag.File = filename
			}
			if !flag.AutoConfirm && utils.FileExist(flag.File) {
				logrus.Println("文件已存在，是否要覆盖?")
				fmt.Print("请确认(y/n): ")
				var input string
				_, err := fmt.Scanln(&input)
				if err != nil {
					return err
				}
				if strings.HasPrefix(strings.ToLower(input), "n") {
					return nil
				}
			}
			txt, err := yamltemplate.GetTemplate(flag.Version)
			if err != nil {
				return err
			}
			logs.Info("初始化文件...")
			f, err := os.Create(flag.File)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = f.WriteString(txt)
			if err != nil {
				return err
			}
			logs.Info("初始化成功")
			return nil
		},
	}
	flags := initCmd.Flags()
	{
		flags.StringVarP(&flag.File, "file", "f", "press.yaml", "文件名")
	}
	{
		flags.StringVarP(&flag.Version, "version", "v", "latest", "版本号")
	}
	{
		flags.BoolVarP(&flag.AutoConfirm, "auto-confirm", "y", false, "自动确认")
	}
	return initCmd
}
