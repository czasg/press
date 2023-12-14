package yamltemplate

import (
	"context"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"net/http"
	"time"
)

type AssertResponse func(response *http.Response) error

var (
	AssertStatusCodeError = errors.New("Assert Status Code Error")
	AssertHeaderError     = errors.New("Assert Header Error")
	AssertBodyError       = errors.New("Assert Body Error")
)

func GetTemplate(version string) (string, error) {
	switch version {
	case "1":
		return NewTemplateV1(), nil
	case "2":
		return NewTemplateV2(), nil
	default:
		return "", fmt.Errorf("unsupport version[%s]", version)
	}
}

type IConfig interface {
	GetVersion() string
	GetSteps() []IStep
}

type IStep interface {
	Print()
	NewRequest(ctx context.Context) (*http.Request, error)
	NewClient() *http.Client
	NewAssert() AssertResponse
	NewStopTimer() *time.Timer
	NewIntervalTicker() *time.Ticker
	NewThreadRampUp(ctx context.Context) func(thread func(ctx context.Context))
}

type BasicConfigVersion struct {
	Version string
}

func Parse(body []byte) (IConfig, error) {
	cfg := BasicConfigVersion{}
	err := yaml.Unmarshal(body, &cfg)
	if err != nil {
		return nil, err
	}
	switch cfg.Version {
	case "1":
		return ParseConfigV1(body)
	case "2":
		return ParseConfigV2(body)
	default:
		return nil, fmt.Errorf("Unsupport Version[%s]", cfg.Version)
	}
}

func ParseVersion(body []byte) (*BasicConfigVersion, error) {
	cfg := BasicConfigVersion{}
	err := yaml.Unmarshal(body, &cfg)
	if err != nil {
		return nil, err
	}
	switch cfg.Version {
	case "1":
	case "2":
	default:
		return nil, err
	}
	return &cfg, nil
}
