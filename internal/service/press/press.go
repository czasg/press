package press

import (
	"context"
	"fmt"
	"github.com/czasg/press/internal/yamltemplate"
	"time"
)

func RunPress(ctx context.Context, cfg yamltemplate.IConfig) error {
	switch cfg.GetVersion() {
	case "1":
	case "2":
	default:
		return fmt.Errorf("Unknown Config")
	}
	return RunPressBySteps(ctx, cfg.GetSteps())
}

func RunPressBySteps(ctx context.Context, steps []yamltemplate.IStep) error {
	for _, step := range steps {
		step.Print()
		pm := &PressManager{
			Stat: &PressStat{},
			Step: step,
		}
		pm.RunPress(ctx)
	}
	return nil
}

type PressManager struct {
	Stat IStat
	Step yamltemplate.IStep
}

func (pm *PressManager) RunPress(ctx context.Context) error {
	ctx1, cancel := context.WithCancel(ctx)
	defer cancel()

	_, err := pm.Step.NewRequest(ctx1)
	if err != nil {
		return err
	}

	threadRampUp := pm.Step.NewThreadRampUp(ctx1)
	go threadRampUp(pm.worker)

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
			snapshot := pm.Stat.Snapshot()
			fmt.Printf("%#v\n", snapshot)
		}
	}
}

func (pm *PressManager) worker(ctx context.Context) {
	pm.Stat.RecordThread()
	assert := pm.Step.NewAssert()
	client := pm.Step.NewClient()
	req, _ := pm.Step.NewRequest(ctx)
	for {
		select {
		case <-ctx.Done():
		default:
		}
		err := func() error {
			defer pm.Stat.RecordTime(time.Now())
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println(err.Error())
				return err
			}
			defer resp.Body.Close()
			return assert(resp)
		}()
		if err != nil {
			pm.Stat.RecordFailure()
		} else {
			pm.Stat.RecordSuccess()
		}
	}
}
