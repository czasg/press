package yamltemplate

import (
	"context"
	"errors"
	"fmt"
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

type IStep interface {
	NewRequest(ctx context.Context) (*http.Request, error)
	NewClient() *http.Client
	NewAssert() AssertResponse
	NewStopTimer() *time.Timer
	NewIntervalTicker() *time.Ticker
}
