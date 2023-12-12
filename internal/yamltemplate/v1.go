package yamltemplate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"net/http"
	"strings"
)

func NewTemplateV1() string {
	return `---
version: "1.0"
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

type ThreadGroupV1 struct {
	Thread       int `json:"thread" yaml:"thread"`
	ThreadRampUp int `json:"threadRampUp" yaml:"threadRampUp"`
	Duration     int `json:"duration" yaml:"duration"`
}

type HttpV1 struct {
	Url     string            `json:"url" yaml:"url"`
	Method  string            `json:"method" yaml:"method"`
	Timeout int               `json:"timeout" yaml:"timeout"`
	Headers map[string]string `json:"headers" yaml:"headers"`
	Body    string            `json:"body" yaml:"body"`
}

func (h *HttpV1) NewRequest(ctx context.Context) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, h.Method, h.Url, bytes.NewBufferString(h.Body))
	if err != nil {
		return nil, err
	}
	for k, v := range h.Headers {
		req.Header.Add(k, v)
	}
	return req, nil
}

type AssertV1 struct {
	StatusCode int                      `json:"statusCode" yaml:"statusCode"`
	Headers    []map[string]string      `json:"headers" yaml:"headers"`
	Body       string                   `json:"body" yaml:"body"`
	JsonMap    []map[string]interface{} `json:"jsonMap" yaml:"jsonMap"`
}

func (a *AssertV1) NewAssert() AssertResponse {
	return func(response *http.Response) error {
		if a.StatusCode > 0 && response.StatusCode != a.StatusCode {
			return AssertStatusCodeError
		}
		for _, headers := range a.Headers {
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
		if a.Body != "" && a.Body != string(body) {
			return AssertBodyError
		}
		if len(a.JsonMap) < 1 {
			return nil
		}
		var m map[string]interface{}
		err = json.Unmarshal(body, &m)
		if err != nil {
			return err
		}
		for _, jsonMap := range a.JsonMap {
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

func ParseConfigV1(body []byte) (*ConfigV1, error) {
	cfg := ConfigV1{}
	err := yaml.Unmarshal(body, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
