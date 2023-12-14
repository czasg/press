package press

import (
	"context"
	"fmt"
	"github.com/czasg/press/internal/yamltemplate"
	"time"
)

func RunPress(ctx context.Context, config yamltemplate.IConfig) error {
	for _, step := range config.GetSteps() {
		step.Print()
		pm := &PressManager{
			Stat:   &PressStat{Step: step},
			Step:   step,
			Config: config,
		}
		err := pm.RunPress(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

type PressManager struct {
	Stat   IStat
	Step   yamltemplate.IStep
	Config yamltemplate.IConfig
}

func (pm *PressManager) RunPress(ctx context.Context) error {
	ctx1, cancel := context.WithCancel(ctx)
	defer func() {
		_ = pm.Close()
		cancel()
	}()
	// check request
	_, err := pm.Step.NewRequest(ctx1)
	if err != nil {
		return err
	}
	// start worker
	go pm.Step.NewThreadRampUp(ctx1)(pm.worker)
	// start stat snapshot
	go pm.Stat.IntervalSnapshotWithHandler(ctx1, NewSnapshotHandler(pm.Config))
	// wait
	select {
	case <-ctx1.Done():
		return ctx1.Err()
	case <-pm.Step.NewStopTimer().C:
		return nil
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

func (pm *PressManager) Close() error {
	return pm.Stat.Close()
}
