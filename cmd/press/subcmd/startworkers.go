package subcmd

import (
	"context"
	"fmt"
	"github.com/czasg/press/internal/service"
	"github.com/czasg/press/third"
	"github.com/spf13/cobra"
	"sync"
	"time"
)

func NewPressStartWorkerCommand(ctx context.Context) *cobra.Command {
	workerCmd := &cobra.Command{
		Use:   "worker",
		Short: "start a press worker",
		Long:  `start a press worker`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := readConfig(cmd)
			if err != nil {
				return err
			}
			rds, err := third.NewRedis(cfg)
			if err != nil {
				return err
			}
			redisKey := fmt.Sprintf("%d", cfg.Metadata.Uid)
			var once sync.Once
			luaScript := `
local redisKey = KEYS[1]

local response_time_min = tonumber(ARGV[1])
local current_response_time_min = tonumber(redis.call('HGET', redisKey, 'ResponseTimeMin'))

if not current_response_time_min or response_time_min < current_response_time_min then
    redis.call('HSET', redisKey, 'ResponseTimeMin', ARGV[1])
end

local response_time_max = tonumber(ARGV[2])
local current_response_time_max = tonumber(redis.call('HGET', redisKey, 'ResponseTimeMax'))

if not current_response_time_max or response_time_max > current_response_time_max then
    redis.call('HSET', redisKey, 'ResponseTimeMax', ARGV[2])
end
`
			snapshotChan := make(chan service.Snapshot, 1)
			snapshotChan <- service.Snapshot{}
			return service.RunPressWithSnapshotHandler(ctx, cfg, func(ctx context.Context, snapshot service.Snapshot) {
				lastSnapshot := <-snapshotChan
				rds.HIncrBy(redisKey, "TotalResponseTime", snapshot.TotalResponseTime-lastSnapshot.TotalResponseTime)
				rds.HIncrBy(redisKey, "TotalRequestCount", snapshot.TotalRequestCount-lastSnapshot.TotalRequestCount)
				rds.HIncrBy(redisKey, "TotalFailureRequestCount", snapshot.TotalFailureRequestCount-lastSnapshot.TotalFailureRequestCount)
				rds.HIncrBy(redisKey, "ThreadNum", snapshot.ThreadNum-lastSnapshot.ThreadNum)
				once.Do(func() {
					rds.HSetNX(redisKey, "StartAt", time.Now().Format(time.RFC3339Nano))
				})
				rds.Eval(luaScript, []string{redisKey}, snapshot.ResponseTimeMin, snapshot.ResponseTimeMax)
				snapshotChan <- snapshot
				fmt.Printf("%#v\n", snapshot)
			})
		},
	}
	cf := workerCmd.Flags()
	{
		cf.StringP("file", "f", "press.yaml", "压力测试配置文件")
	}
	return workerCmd
}
