package service

import (
	"context"
	"fmt"
	"github.com/czasg/press/internal/config"
	"github.com/sirupsen/logrus"
	"time"
)

func RunPress(ctx context.Context, cfg *config.Config) error {
	return RunPressWithSnapshotHandler(ctx, cfg, snapshotLogHandler)
}

func RunPressWithSnapshotHandler(ctx context.Context, cfg *config.Config, handler SnapshotHandler) error {
	logWithConfig(cfg)
	for _, step := range cfg.Steps {
		logWithStep(step)
		pm := PressManager{
			Stat:            &Stat{},
			StepManager:     NewStepManager(step),
			SnapshotHandler: handler,
		}
		err := pm.RunPress(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

type PressManager struct {
	Stat            *Stat
	StepManager     *StepManager
	SnapshotHandler SnapshotHandler
}

func (pm *PressManager) RunPress(ctx context.Context) error {
	ctx1, cancel := context.WithCancel(ctx)
	defer cancel()
	go pm.IntervalSnapshot(ctx1)
	go pm.WorkerGroup(ctx1)
	select {
	case <-ctx1.Done():
		return nil
	case <-pm.StepManager.NewStopTimer().C:
		return nil
	}
}

func (pm *PressManager) IntervalSnapshot(ctx context.Context) error {
	ticker := pm.StepManager.NewIntervalTicker()
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			pm.SnapshotHandler(ctx, pm.Stat.Snapshot())
		}
	}
}

func (pm *PressManager) WorkerGroup(ctx context.Context) error {
	interval := time.Second * time.Duration(pm.StepManager.ThreadGroup.ThreadRampUp/pm.StepManager.ThreadGroup.Thread)
	timer := time.NewTimer(0)
	defer timer.Stop()
	for i := 0; i < pm.StepManager.ThreadGroup.Thread; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			go NewWorker(pm.Stat, pm.StepManager).Start(ctx)
			timer.Reset(interval)
		}
	}
	return nil
}

func (pm *PressManager) Close() error {
	return pm.Stat.Close()
}

func (pm *PressManager) Check(ctx context.Context) error {
	_, err := pm.StepManager.NewRequest(ctx)
	return err
}

func logWithConfig(cfg *config.Config) {
	logrus.WithField("Version", cfg.Version).Info("检测到当前版本")
	logrus.WithField("User", cfg.Metadata.Name).Info("检测到当前用户")
	logrus.WithField("Uid", cfg.Metadata.Uid).Info("检测到任务编码")
	fmt.Println("----------------------------------------------------------------------------")
}

func logWithStep(step config.Steps) {
	logrus.Println("任务启动")
	logrus.Printf("名称：%v", step.Name)
	logrus.Printf("线程数：%v", step.ThreadGroup.Thread)
	logrus.Printf("线程唤醒时间：%vs", step.ThreadGroup.ThreadRampUp)
	logrus.Printf("持续时间：%vs", step.ThreadGroup.Duration)
	logrus.Printf("日志输出间隔：%vs", step.LogInterval)
}
