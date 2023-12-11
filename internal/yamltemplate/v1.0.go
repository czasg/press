package yamltemplate

import "gopkg.in/yaml.v2"

func NewTemplateV1() string {
	return `---
version: "1.0"
metadata:
  name: "press"
  clusterEnable: false
  clusterRedis:
    host: ""
    port: 6379
    db: 0
    pwd: ""
steps:
  - name: "压力测试"
    logInterval: 1
    threadGroup:
      thread: 2
      threadRampUp: 1
      duration: 10
    http:
      url: "http://www.baidu.com"
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

func ParseConfigV1(body []byte) (*ConfigV1, error) {
	cfg := ConfigV1{}
	err := yaml.Unmarshal(body, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

type ConfigV1 struct {
	Version  string     `json:"version" yaml:"version"`
	Metadata MetadataV1 `json:"metadata" yaml:"metadata"`
	Steps    []StepsV1  `json:"steps" yaml:"steps"`
}
type MetadataV1 struct {
	Name          string         `json:"name" yaml:"name"`
	ClusterEnable bool           `json:"clusterEnable" yaml:"clusterEnable"`
	ClusterRedis  ClusterRedisV1 `json:"clusterRedis" yaml:"clusterRedis"`
}

type ClusterRedisV1 struct {
	Host string `json:"host" yaml:"host"`
	Port int    `json:"port" yaml:"port"`
	DB   int    `json:"db" yaml:"db"`
	Pwd  string `json:"pwd" yaml:"pwd"`
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

type AssertV1 struct {
	StatusCode int                      `json:"statusCode" yaml:"statusCode"`
	Headers    []map[string]string      `json:"headers" yaml:"headers"`
	Body       string                   `json:"body" yaml:"body"`
	JsonMap    []map[string]interface{} `json:"jsonMap" yaml:"jsonMap"`
}
