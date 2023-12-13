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
}

func (w *Worker) RunPress(ctx context.Context, step interface{}) error {
	switch (step).(type) {
	case yamltemplate.StepsV1:
		return w.RunPressV1(ctx, step.(yamltemplate.StepsV1))
	case yamltemplate.StepsV2:
		return nil
	default:
		return fmt.Errorf("Unsupport Step[%v]", step)
	}
}

func (w *Worker) RunPressV1(ctx context.Context, step yamltemplate.StepsV1) error {
	req, err := step.Http.NewRequest(ctx)
	if err != nil {
		return err
	}
	client := step.Http.NewClient(ctx)
	assert := step.Assert.NewAssert()

	go func() {
		interval := time.Second * time.Duration(step.ThreadGroup.ThreadRampUp) / time.Duration(step.ThreadGroup.Thread)
		for i := 0; i < step.ThreadGroup.Thread; i++ {
			select {
			case <-ctx.Done():
				return
			default:
				go w.worker(ctx, req, client, assert)
				time.Sleep(interval)
			}
		}
	}()

	return nil
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
