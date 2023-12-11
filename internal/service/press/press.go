package press

import (
	"context"
	"fmt"
	"github.com/czasg/press/internal/yamltemplate"
	"github.com/sirupsen/logrus"
)

func RunPress(ctx context.Context, cfg interface{}) error {
	switch cfg.(type) {
	case yamltemplate.ConfigV1:
		return RunPressV1(ctx, cfg.(yamltemplate.ConfigV1))
	default:
		return fmt.Errorf("Unknown Config")
	}
}

func RunPressV1(ctx context.Context, cfg yamltemplate.ConfigV1) error {
	logrus.WithField("Version", cfg.Version).Info("检测到当前版本")
	logrus.WithField("User", cfg.Metadata.Name).Info("检测到当前用户")
	for index, step := range cfg.Steps {
		logrus.Info("#########################")
		logrus.Printf("###### 任务[%v]开始 ######", index)
		logrus.Info("#########################")
		logrus.Printf("名称：%v", step.Name)
		logrus.Printf("线程数：%v", step.ThreadGroup.Thread)
		logrus.Printf("线程唤醒时间：%vs", step.ThreadGroup.ThreadRampUp)
		logrus.Printf("持续时间：%vs", step.ThreadGroup.Duration)
		logrus.Printf("日志输出间隔：%vs", step.LogInterval)
	}
	logrus.Info("##########################")
	logrus.Info("###### 压力测试结束 ######")
	logrus.Info("##########################")
	return nil
}
