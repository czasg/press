package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/czasg/press/internal/config"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	AssertStatusCodeError = errors.New("Assert Status Code Error")
	AssertHeaderError     = errors.New("Assert Header Error")
	AssertBodyError       = errors.New("Assert Body Error")
)

func NewStepManager(steps config.Steps) *StepManager {
	return &StepManager{Steps: steps}
}

type StepManager struct {
	config.Steps
}

func (sm *StepManager) NewRequest(ctx context.Context) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, sm.Http.Method, sm.Http.Url, bytes.NewBufferString(sm.Http.Body))
	if err != nil {
		return nil, err
	}
	for k, v := range sm.Http.Headers {
		req.Header.Add(k, v)
	}
	return req, nil
}

func (sm *StepManager) NewStopTimer() *time.Timer {
	return time.NewTimer(time.Second * time.Duration(sm.ThreadGroup.Duration))
}

func (sm *StepManager) NewIntervalTicker() *time.Ticker {
	return time.NewTicker(time.Second * time.Duration(sm.LogInterval))
}

func (sm *StepManager) NewClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: !sm.Http.Keepalive,
		},
		Timeout: time.Duration(float64(time.Second) * sm.Http.Timeout),
	}
}

func (sm *StepManager) NewAssert() func(response *http.Response) error {
	return func(response *http.Response) error {
		if sm.Assert.StatusCode > 0 && response.StatusCode != sm.Assert.StatusCode {
			return AssertStatusCodeError
		}
		for _, headers := range sm.Assert.Headers {
			for k, v := range headers {
				if !strings.EqualFold(response.Header.Get(k), v) {
					return AssertHeaderError
				}
			}
		}
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}
		if sm.Assert.Body != "" && sm.Assert.Body != string(body) {
			return AssertBodyError
		}
		if len(sm.Assert.JsonMap) < 1 {
			return nil
		}
		var m map[string]interface{}
		err = json.Unmarshal(body, &m)
		if err != nil {
			return err
		}
		for _, jsonMap := range sm.Assert.JsonMap {
			for k, v := range jsonMap {
				v1, ok := m[k]
				if !ok {
					return AssertBodyError
				}
				if !strings.EqualFold(
					fmt.Sprintf("%v", v), fmt.Sprintf("%v", v1),
				) {
					return AssertBodyError
				}
			}
		}
		return nil
	}
}
