package config

import (
	"github.com/czasg/snow"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Version  string   `json:"version" yaml:"version"`
	Metadata Metadata `json:"metadata" yaml:"metadata"`
	Steps    []Steps  `json:"steps" yaml:"steps"`
}

type Metadata struct {
	Name        string      `json:"name" yaml:"name"`
	Uid         int64       `json:"uid"`
	Annotations Annotations `json:"annotations" yaml:"annotations"`
}

type Annotations struct {
	PressClusterBroker               string `json:"pressClusterBroker" yaml:"press.cluster.broker"`
	PressClusterBrokerEnabled        bool   `json:"pressClusterBrokerEnabled" yaml:"press.cluster.broker/enabled"`
	PressClusterBrokerRedisUrl       string `json:"pressClusterBrokerRedisUrl" yaml:"press.cluster.broker/redis.url"`
	PressClusterBrokerRedisPbWorker  string `json:"pressClusterBrokerRedisPbWorker" yaml:"press.cluster.broker/redis.pb.worker"`
	PressClusterBrokerRedisPbManager string `json:"pressClusterBrokerRedisPbManager" yaml:"press.cluster.broker/redis.pb.manager"`
}

type Steps struct {
	Name        string      `json:"name" yaml:"name"`
	LogInterval int         `json:"logInterval" yaml:"logInterval"`
	ThreadGroup ThreadGroup `json:"threadGroup" yaml:"threadGroup"`
	Http        Http        `json:"http" yaml:"http"`
	Assert      Assert      `json:"assert" yaml:"assert"`
}

type ThreadGroup struct {
	Thread       int `json:"thread" yaml:"thread"`
	ThreadRampUp int `json:"threadRampUp" yaml:"threadRampUp"`
	Duration     int `json:"duration" yaml:"duration"`
}

type Http struct {
	Url       string            `json:"url" yaml:"url"`
	Method    string            `json:"method" yaml:"method"`
	Timeout   float64           `json:"timeout" yaml:"timeout"`
	Keepalive bool              `json:"keepalive" yaml:"keepalive"`
	Headers   map[string]string `json:"headers" yaml:"headers"`
	Body      string            `json:"body" yaml:"body"`
}

type Assert struct {
	StatusCode int                      `json:"statusCode" yaml:"statusCode"`
	Headers    []map[string]string      `json:"headers" yaml:"headers"`
	Body       string                   `json:"body" yaml:"body"`
	JsonMap    []map[string]interface{} `json:"jsonMap" yaml:"jsonMap"`
}

func NewConfigTemplate() string {
	return `---
version: "1"
metadata:
  name: "press"
  annotations:
    press.cluster.broker: redis
    press.cluster.broker/enabled: false
    press.cluster.broker/redis.url: "redis://:@localhost:6379/0"
    press.cluster.broker/redis.pb.worker: "press-test-worker"
    press.cluster.broker/redis.pb.manager: "press-test-manager"
steps:
  - name: "press test"
    logInterval: 1
    threadGroup:
      thread: 1
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

func Parse(body []byte) (*Config, error) {
	cfg := Config{}
	err := yaml.Unmarshal(body, &cfg)
	if err != nil {
		return nil, err
	}
	cfg.Metadata.Uid = snow.Next()
	return &cfg, nil
}
