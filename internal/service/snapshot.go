package service

import (
	"context"
	"fmt"
	"github.com/czasg/press/internal/yamltemplate"
)

type SnapshotHandler func(ctx context.Context, snapshot Snapshot)

type Snapshot struct {
	Throughput               int64
	ThroughputMean           int64
	ResponseTimeMin          int64
	ResponseTimeMean         int64
	ResponseTimeMax          int64
	TotalFailureRequestCount int64
	TotalRequestCount        int64
	ThreadNum                int64
}

func NewSnapshotHandler(config yamltemplate.IConfig) SnapshotHandler {
	switch config.GetVersion() {
	case "1":
		return SnapshotLogHandler
	default:
		return SnapshotLogHandler
	}
}

func SnapshotLogHandler(ctx context.Context, snapshot Snapshot) {
	fmt.Printf("%#v\n", snapshot)
}

func SnapshotRedisHandler(ctx context.Context, snapshot Snapshot) {
	fmt.Printf("%#v\n", snapshot)
}
