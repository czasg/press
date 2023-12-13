package press

import (
	"context"
	"fmt"
	"github.com/czasg/press/internal/yamltemplate"
	"net/http"
	"time"
)

type Worker struct {
	Stat IStat
	Step yamltemplate.IStep
}

func (w *Worker) RunPress(ctx context.Context) error {
	ctx1, cancel := context.WithCancel(ctx)
	defer cancel()

	assert := w.Step.NewAssert()
	client := w.Step.NewClient()
	req, err := w.Step.NewRequest(ctx1)
	if err != nil {
		return err
	}
	workerThread := func(ctx context.Context) {
		w.worker(ctx, req, client, assert)
	}

	threadRampUp := w.Step.NewThreadRampUp(ctx1)
	go threadRampUp(workerThread)

	currentSnapshotTime := time.Now()
	intervalTicker := w.Step.NewIntervalTicker()
	stopTimer := w.Step.NewStopTimer()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-stopTimer.C:
			w.Stat.Close()
			cancel()
			return nil
		case <-intervalTicker.C:
			snapshot := w.Stat.Snapshot(currentSnapshotTime)
			currentSnapshotTime = time.Now()
			fmt.Printf("%#v\n", snapshot)
		}
	}
}

func (w *Worker) worker(
	ctx context.Context,
	req *http.Request,
	client *http.Client,
	assertResponse yamltemplate.AssertResponse,
) {
	w.Stat.RecordThread()
	for {
		select {
		case <-ctx.Done():
		default:
		}
		err := func() error {
			defer w.Stat.RecordTime(time.Now())
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
			w.Stat.RecordFailure()
		} else {
			w.Stat.RecordSuccess()
		}
	}
}
