package press

import (
	"context"
	"fmt"
	"github.com/czasg/press/internal/yamltemplate"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
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
		pm := &PressManager{
			Stat: &PressStat{},
		}
		pm.RunPressV1(ctx, step)
	}
	logrus.Info("##########################")
	logrus.Info("###### 压力测试结束 ######")
	logrus.Info("##########################")
	return nil
}

type PressManager struct {
	Stat IStat
}

func (pm *PressManager) RunPressV1(ctx context.Context, step yamltemplate.StepsV1) error {
	ctx1, cancel := context.WithCancel(ctx)
	defer cancel()

	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
	req, err := step.Http.NewRequest(ctx1)
	if err != nil {
		return err
	}
	assert := step.Assert.NewAssert()

	go func() {
		interval := time.Second * time.Duration(step.ThreadGroup.ThreadRampUp) / time.Duration(step.ThreadGroup.Thread)
		for i := 0; i < step.ThreadGroup.Thread; i++ {
			select {
			case <-ctx1.Done():
				return
			default:
				go pm.worker(ctx1, req, client, assert)
				time.Sleep(interval)
			}
		}
	}()

	closeTimer := time.NewTimer(time.Second * time.Duration(step.ThreadGroup.Duration))
	intervalTicker := time.NewTicker(time.Second * time.Duration(step.LogInterval))
	currentSnapshotTime := time.Now()
	for {
		select {
		case <-ctx1.Done():
			return nil
		case <-closeTimer.C:
			pm.Stat.Close()
			cancel()
			return nil
		case <-intervalTicker.C:
			snapshot := pm.Stat.Snapshot(currentSnapshotTime)
			currentSnapshotTime = time.Now()
			fmt.Printf("%#v\n", snapshot)
		}
	}
}

//func (pm *PressManager) startWorkers(ctx context.Context, step yamltemplate.StepsV1) error {
//	client := &http.Client{
//		Transport: &http.Transport{
//			DisableKeepAlives: true,
//		},
//	}
//	req, err := step.Http.NewRequest(ctx)
//	if err != nil {
//		return err
//	}
//	assert := step.Assert.NewAssert()
//
//	go func() {
//		interval := time.Second * time.Duration(step.ThreadGroup.ThreadRampUp) / time.Duration(step.ThreadGroup.Thread)
//		for i := 0; i < step.ThreadGroup.Thread; i++ {
//			select {
//			case <-ctx.Done():
//				return
//			default:
//				go pm.worker(ctx, req, client, assert)
//				time.Sleep(interval)
//			}
//		}
//	}()
//
//	return nil
//}

func (pm *PressManager) worker(
	ctx context.Context,
	req *http.Request,
	client *http.Client,
	assertResponse yamltemplate.AssertResponse,
) {
	pm.Stat.RecordThread()
	for {
		select {
		case <-ctx.Done():
		default:
		}
		err := func() error {
			defer pm.Stat.RecordTime(time.Now())
			timeoutCtx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()
			resp, err := client.Do(req.WithContext(timeoutCtx))
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			return assertResponse(resp)
		}()
		if err != nil {
			pm.Stat.RecordFailure()
		} else {
			pm.Stat.RecordSuccess()
		}
	}
}
