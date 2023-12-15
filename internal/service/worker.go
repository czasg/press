package service

import (
	"context"
	"time"
)

func NewWorker(stat *Stat, stepManager *StepManager) *Worker {
	return &Worker{
		Stat:        stat,
		StepManager: stepManager,
	}
}

type Worker struct {
	Stat        *Stat
	StepManager *StepManager
}

func (w *Worker) Start(ctx context.Context) {
	w.Stat.RecordThread()
	assert := w.StepManager.NewAssert()
	client := w.StepManager.NewClient()
	req, _ := w.StepManager.NewRequest(ctx)
	for {
		select {
		case <-ctx.Done():
		default:
		}
		err := func() error {
			defer w.Stat.RecordTime(time.Now())
			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			return assert(resp)
		}()
		if err != nil {
			w.Stat.RecordFailure()
		} else {
			w.Stat.RecordSuccess()
		}
	}
}
