package service

import (
	"context"
	"fmt"
)

type SnapshotHandler func(ctx context.Context, snapshot Snapshot)

type Snapshot struct {
	Throughput               int64
	ThroughputMean           int64
	ResponseTimeMin          int64
	ResponseTimeMean         int64
	ResponseTimeMax          int64
	TotalResponseTime        int64
	TotalFailureRequestCount int64
	TotalRequestCount        int64
	ThreadNum                int64
}

func snapshotLogHandler(ctx context.Context, snapshot Snapshot) {
	fmt.Printf("%#v\n", snapshot)
}
