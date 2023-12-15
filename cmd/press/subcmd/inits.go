package subcmd

import (
	"context"
	"fmt"
	"github.com/czasg/press/internal/config"
	"github.com/czasg/press/internal/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

type PressInitCommandFlag struct {
	File        string `json:"file" yaml:"file"`
	AutoConfirm bool   `json:"autoConfirm" yaml:"autoConfirm"`
}

func NewPressInitCommand(ctx context.Context) *cobra.Command {
	flag := PressInitCommandFlag{}
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "init a press template yaml",
		Long: `init a press template yaml, version list:
- 1: single pressure test.
- 2: cluster pressure test.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !filepath.IsAbs(flag.File) {
				filename, err := filepath.Abs(flag.File)
				if err != nil {
					return err
				}
				logrus.Infof("fixed file absolute filepath [%s]->[%s]", flag.File, filename)
				flag.File = filename
			}
			if !flag.AutoConfirm && utils.FileExist(flag.File) {
				logrus.Println("file already exists, override?")
				fmt.Print("please(y/n): ")
				var input string
				_, err := fmt.Scanln(&input)
				if err != nil {
					return err
				}
				if strings.HasPrefix(strings.ToLower(input), "n") {
					return nil
				}
			}
			f, err := os.Create(flag.File)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = f.WriteString(config.NewConfigTemplate())
			if err != nil {
				return err
			}
			logrus.Info("init success!")
			return nil
		},
	}
	flags := initCmd.Flags()
	{
		flags.StringVarP(&flag.File, "file", "f", "press.yaml", "filename")
	}
	{
		flags.BoolVarP(&flag.AutoConfirm, "auto-confirm", "y", false, "auto confirm `yes` if file exists.")
	}
	return initCmd
}
