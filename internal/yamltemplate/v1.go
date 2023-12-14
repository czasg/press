package yamltemplate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	_ IConfig = (*ConfigV1)(nil)
	_ IStep   = (*StepsV1)(nil)
)

func NewTemplateV1() string {
	return `---
version: "1"
metadata:
  name: "press"
steps:
  - name: "压力测试"
    logInterval: 1
    threadGroup:
      thread: 2
      threadRampUp: 1
      duration: 10
    http:
      url: "http://localhost:8080"
      method: "GET"
      timeout: 10
      keepalive: false
    #      headers:
    #        content-type: "application/json"
    #      body: |
    #        {
    #          "hello":"press"
    #        }
    assert:
      statusCode: 200
#      headers:
#        - auth: "authKey"
#      body: ""
#      jsonMap:  # 仅支持获取 Map 第一层
#        - errCode: 0
#        - status: true
`
}

type ConfigV1 struct {
	Version  string     `json:"version" yaml:"version"`
	Metadata MetadataV1 `json:"metadata" yaml:"metadata"`
	Steps    []StepsV1  `json:"steps" yaml:"steps"`
}

func (c *ConfigV1) GetVersion() string {
	return c.Version
}

func (c *ConfigV1) GetSteps() []IStep {
	ss := []IStep{}
	for _, step := range c.Steps {
		ss = append(ss, &step)
	}
	return ss
}

type MetadataV1 struct {
	Name string `json:"name" yaml:"name"`
}

type StepsV1 struct {
	Name        string        `json:"name" yaml:"name"`
	LogInterval int           `json:"logInterval" yaml:"logInterval"`
	ThreadGroup ThreadGroupV1 `json:"threadGroup" yaml:"threadGroup"`
	Http        HttpV1        `json:"http" yaml:"http"`
	Assert      AssertV1      `json:"assert" yaml:"assert"`
}

func (s *StepsV1) Print() {
	logrus.Printf("Name：%v", s.Name)
	logrus.Printf("Thread：%v", s.ThreadGroup.Thread)
	logrus.Printf("ThreadRampUp：%vs", s.ThreadGroup.ThreadRampUp)
	logrus.Printf("Duration：%vs", s.ThreadGroup.Duration)
	logrus.Printf("LogInterval：%vs", s.LogInterval)
	logrus.Printf("Url：%v", s.Http.Url)
}

func (s *StepsV1) NewRequest(ctx context.Context) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, s.Http.Method, s.Http.Url, bytes.NewBufferString(s.Http.Body))
	if err != nil {
		return nil, err
	}
	for k, v := range s.Http.Headers {
		req.Header.Add(k, v)
	}
	return req, nil
}

func (s *StepsV1) NewClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: !s.Http.Keepalive,
		},
		Timeout: time.Second * time.Duration(s.Http.Timeout),
	}
}

func (s *StepsV1) NewStopTimer() *time.Timer {
	return time.NewTimer(time.Second * time.Duration(s.ThreadGroup.Duration))
}

func (s *StepsV1) NewIntervalTicker() *time.Ticker {
	return time.NewTicker(time.Second * time.Duration(s.LogInterval))
}

func (s *StepsV1) NewThreadRampUp(ctx context.Context) func(thread func(ctx context.Context)) {
	var once sync.Once
	return func(thread func(ctx context.Context)) {
		once.Do(func() {
			interval := time.Second * time.Duration(s.ThreadGroup.ThreadRampUp) / time.Duration(s.ThreadGroup.Thread)
			for i := 0; i < s.ThreadGroup.Thread; i++ {
				select {
				case <-ctx.Done():
					return
				default:
					go thread(ctx)
					time.Sleep(interval)
				}
			}
		})
	}
}

func (s *StepsV1) NewAssert() AssertResponse {
	return func(response *http.Response) error {
		if s.Assert.StatusCode > 0 && response.StatusCode != s.Assert.StatusCode {
			return AssertStatusCodeError
		}
		for _, headers := range s.Assert.Headers {
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
		if s.Assert.Body != "" && s.Assert.Body != string(body) {
			return AssertBodyError
		}
		if len(s.Assert.JsonMap) < 1 {
			return nil
		}
		var m map[string]interface{}
		err = json.Unmarshal(body, &m)
		if err != nil {
			return err
		}
		for _, jsonMap := range s.Assert.JsonMap {
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

type ThreadGroupV1 struct {
	Thread       int `json:"thread" yaml:"thread"`
	ThreadRampUp int `json:"threadRampUp" yaml:"threadRampUp"`
	Duration     int `json:"duration" yaml:"duration"`
}

//func (t ThreadGroupV1) NewThreadRampUp(ctx context.Context) func(ctx context.Context, thread func()) {
//	var once sync.Once
//	return func(ctx context.Context, thread func()) {
//		once.Do(func() {
//			go func() {
//				interval := time.Second * time.Duration(t.ThreadRampUp) / time.Duration(t.Thread)
//				for i := 0; i < t.Thread; i++ {
//					select {
//					case <-ctx.Done():
//						return
//					default:
//						go thread()
//						time.Sleep(interval)
//					}
//				}
//			}()
//		})
//	}
//}

type HttpV1 struct {
	Url       string            `json:"url" yaml:"url"`
	Method    string            `json:"method" yaml:"method"`
	Timeout   int               `json:"timeout" yaml:"timeout"`
	Keepalive bool              `json:"keepalive" yaml:"keepalive"`
	Headers   map[string]string `json:"headers" yaml:"headers"`
	Body      string            `json:"body" yaml:"body"`
}

//func (h *HttpV1) NewRequest(ctx context.Context) (*http.Request, error) {
//	req, err := http.NewRequestWithContext(ctx, h.Method, h.Url, bytes.NewBufferString(h.Body))
//	if err != nil {
//		return nil, err
//	}
//	for k, v := range h.Headers {
//		req.Header.Add(k, v)
//	}
//	return req, nil
//}

//func (h *HttpV1) NewClient(ctx context.Context) *http.Client {
//	return &http.Client{
//		Transport: &http.Transport{
//			DisableKeepAlives: !h.Keepalive,
//		},
//		Timeout: time.Second * time.Duration(h.Timeout),
//	}
//}

type AssertV1 struct {
	StatusCode int                      `json:"statusCode" yaml:"statusCode"`
	Headers    []map[string]string      `json:"headers" yaml:"headers"`
	Body       string                   `json:"body" yaml:"body"`
	JsonMap    []map[string]interface{} `json:"jsonMap" yaml:"jsonMap"`
}

//func (a *AssertV1) NewAssert() AssertResponse {
//	return func(response *http.Response) error {
//		if a.StatusCode > 0 && response.StatusCode != a.StatusCode {
//			return AssertStatusCodeError
//		}
//		for _, headers := range a.Headers {
//			for k, v := range headers {
//				if !strings.EqualFold(response.Header.Get(k), v) {
//					return AssertHeaderError
//				}
//			}
//		}
//		body, err := io.ReadAll(response.Body)
//		if err != nil {
//			return err
//		}
//		if a.Body != "" && a.Body != string(body) {
//			return AssertBodyError
//		}
//		if len(a.JsonMap) < 1 {
//			return nil
//		}
//		var m map[string]interface{}
//		err = json.Unmarshal(body, &m)
//		if err != nil {
//			return err
//		}
//		for _, jsonMap := range a.JsonMap {
//			for k, v := range jsonMap {
//				v1, ok := m[k]
//				if !ok {
//					return AssertBodyError
//				}
//				if !strings.EqualFold(
//					fmt.Sprintf("%v", v), fmt.Sprintf("%v", v1),
//				) {
//					return AssertBodyError
//				}
//			}
//		}
//		return nil
//	}
//}

func ParseConfigV1(body []byte) (*ConfigV1, error) {
	cfg := ConfigV1{}
	err := yaml.Unmarshal(body, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func TransformV1(cfg IConfig) (*ConfigV1, error) {
	body, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	var v1 ConfigV1
	err = json.Unmarshal(body, &v1)
	if err != nil {
		return nil, err
	}
	return &v1, nil
}
