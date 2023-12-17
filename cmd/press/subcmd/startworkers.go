package subcmd

import (
	"context"
	"fmt"
	"github.com/czasg/press/internal/config"
	"github.com/czasg/press/internal/service"
	"github.com/czasg/press/third"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
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
			channelName := cfg.Metadata.Annotations.PressClusterBrokerRedisPbWorker
			sub := rds.WithContext(ctx).Subscribe(channelName)
			for {
				message, err := sub.ReceiveMessage()
				if err != nil {
					continue
				}
				newCfg, err := config.Parse([]byte(message.Payload))
				if err != nil {
					continue
				}
				ctx1, cancel := context.WithCancel(ctx)
				func() {
					defer cancel()
					go func() {
						closeSub := rds.WithContext(ctx1).Subscribe(fmt.Sprintf("%s-%d", channelName, newCfg.Metadata.Uid))
						_, _ = closeSub.ReceiveMessage()
						cancel()
					}()
					defer func() {
						rds.Publish(fmt.Sprintf("%d-closed", newCfg.Metadata.Uid), "closed")
					}()
					rds.RPush(fmt.Sprintf("%d-%d", newCfg.Metadata.Uid, cfg.Metadata.Uid))
					logrus.WithFields(logrus.Fields{
						"Uid": newCfg.Metadata.Uid,
					}).Info("检测到压测任务")
					handler := NewSnapshotRedisHandler(rds, newCfg)
					err = service.RunPressWithSnapshotHandler(ctx1, newCfg, handler)
					if err != nil {
						return
					}
				}()
				<-ctx1.Done()
			}
		},
	}
	cf := workerCmd.Flags()
	{
		cf.StringP("file", "f", "press.yaml", "压力测试配置文件")
	}
	return workerCmd
}

const luaScript = `
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

func NewSnapshotRedisHandler(rds *redis.Client, cfg *config.Config) service.SnapshotHandler {
	var (
		once         sync.Once
		redisKey     = fmt.Sprintf("%d", cfg.Metadata.Uid)
		snapshotChan = make(chan service.Snapshot, 1)
	)
	snapshotChan <- service.Snapshot{}
	return func(ctx context.Context, snapshot service.Snapshot) {
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
	}
}
