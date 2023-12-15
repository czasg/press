package service

import (
	"context"
	"fmt"
	"github.com/czasg/press/internal/config"
	"time"
)

func RunPress(ctx context.Context, config *config.Config) error {
	for _, step := range config.Steps {
		pm := PressManager{
			Stat:        &Stat{},
			StepManager: NewStepManager(step),
		}
		pm.RunPress(ctx)
	}
	return nil
}

type PressManager struct {
	Stat        *Stat
	StepManager *StepManager
}

func (pm *PressManager) RunPress(ctx context.Context) error {
	ctx1, cancel := context.WithCancel(ctx)
	defer cancel()
	go pm.IntervalSnapshot(ctx1)
	go pm.WorkerGroup(ctx1)
	select {
	case <-ctx1.Done():
		return ctx1.Err()
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
			return ctx.Err()
		case <-ticker.C:
			snapshot := pm.Stat.Snapshot()
			fmt.Printf("%#v\n", snapshot)
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
