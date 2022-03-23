package basis

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
)

func MustParseConfig(cfgFilePath string) *Config {
	log.Printf("检测到配置文件: %s\n", cfgFilePath)
	f, err := os.Open(cfgFilePath)
	if err != nil {
		log.Fatalf("打开配置文件异常：%v", err)
	}
	cfgBody, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalf("读取配置文件异常：%v", err)
	}
	var cfg Config
	err = yaml.Unmarshal(cfgBody, &cfg)
	if err != nil {
		log.Fatalf("解析配置文件异常：%v", err)
	}
	return &cfg
}

type Config struct {
	Version  string   `yaml:"version"`
	Metadata Metadata `yaml:"metadata"`
	Steps    []Steps  `yaml:"steps"`
}

type Metadata struct {
	Name string `yaml:"name"`
}

type Steps struct {
	Name        string      `yaml:"name"`
	LogInterval int         `yaml:"logInterval"`
	ThreadGroup ThreadGroup `yaml:"threadGroup"`
	Http        Http        `yaml:"http"`
	Assert      Assert      `yaml:"assert"`
	Output      Output      `yaml:"output"`
}

type ThreadGroup struct {
	Thread       int `yaml:"thread"`
	ThreadRampUp int `yaml:"threadRampUp"`
	Duration     int `yaml:"duration"`
}

type Http struct {
	Url     string            `yaml:"url"`
	Method  string            `yaml:"method"`
	Headers map[string]string `yaml:"headers"`
	Body    string            `yaml:"body"`
}

type Assert struct {
	StatusCode int                      `yaml:"statusCode"`
	Headers    []map[string]string      `yaml:"headers"`
	Body       string                   `yaml:"body"`
	JsonMap    []map[string]interface{} `yaml:"jsonMap"`
}

type Output struct {
	Path string
}
