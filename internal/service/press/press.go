package press

import (
	"context"
	"fmt"
	"github.com/czasg/press/internal/yamltemplate"
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
	for _, step := range cfg.Steps {
		//logrus.Info("#########################")
		//logrus.Printf("######## task[%v] ########", index)
		//logrus.Info("#########################")
		step.Print()
		pm := &PressManager{
			Stat: &PressStat{},
			Step: &step,
		}
		pm.RunPress(ctx)
	}
	//logrus.Info("##########################")
	//logrus.Info("######## over ########")
	//logrus.Info("##########################")
	return nil
}

type PressManager struct {
	Stat IStat
	Step yamltemplate.IStep
}

//func (pm *PressManager) RunPressV1(ctx context.Context, step yamltemplate.StepsV1) error {
//	ctx1, cancel := context.WithCancel(ctx)
//	defer cancel()
//
//	req, err := step.Http.NewRequest(ctx1)
//	if err != nil {
//		return err
//	}
//	client := step.Http.NewClient(ctx1)
//	assert := step.Assert.NewAssert()
//
//	go func() {
//		interval := time.Second * time.Duration(step.ThreadGroup.ThreadRampUp) / time.Duration(step.ThreadGroup.Thread)
//		for i := 0; i < step.ThreadGroup.Thread; i++ {
//			select {
//			case <-ctx1.Done():
//				return
//			default:
//				go pm.worker(ctx1, req, client, assert)
//				time.Sleep(interval)
//			}
//		}
//	}()
//
//	closeTimer := time.NewTimer(time.Second * time.Duration(step.ThreadGroup.Duration))
//	intervalTicker := time.NewTicker(time.Second * time.Duration(step.LogInterval))
//	currentSnapshotTime := time.Now()
//	for {
//		select {
//		case <-ctx1.Done():
//			return nil
//		case <-closeTimer.C:
//			pm.Stat.Close()
//			cancel()
//			return nil
//		case <-intervalTicker.C:
//			snapshot := pm.Stat.Snapshot(currentSnapshotTime)
//			currentSnapshotTime = time.Now()
//			fmt.Printf("%#v\n", snapshot)
//		}
//	}
//}

func (pm *PressManager) RunPress(ctx context.Context) error {
	ctx1, cancel := context.WithCancel(ctx)
	defer cancel()

	assert := pm.Step.NewAssert()
	client := pm.Step.NewClient()
	req, err := pm.Step.NewRequest(ctx1)
	if err != nil {
		return err
	}
	workerThread := func(ctx context.Context) {
		pm.worker(ctx, req, client, assert)
	}

	threadRampUp := pm.Step.NewThreadRampUp(ctx1)
	go threadRampUp(workerThread)

	currentSnapshotTime := time.Now()
	intervalTicker := pm.Step.NewIntervalTicker()
	stopTimer := pm.Step.NewStopTimer()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-stopTimer.C:
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
