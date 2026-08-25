package service

import (
	"context"
	"fmt"
	"time"

	"ebike-device-openapi-go/internal/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const luopingReplayKeyPrefix = "ecu_replay_"

var actionRdb *redis.Client

// SetActionRedis injects Redis used by luoping sync-command waiters.
func SetActionRedis(rdb *redis.Client) {
	actionRdb = rdb
}

func luopingReplayKey(tid string) string {
	return luopingReplayKeyPrefix + tid
}

// waitLuopingReply polls Redis for ecu_replay_{jobId} written when a wild reply
// (header cmd=0) arrives on brpt and matches the pending seq→jobId mapping.
func waitLuopingReply(tid string, timeout time.Duration) (string, error) {
	if tid == "" {
		return "", fmt.Errorf("tid is empty")
	}
	if actionRdb == nil {
		return "", fmt.Errorf("redis not initialized for luoping sync wait")
	}
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	deadline := time.Now().Add(timeout)
	key := luopingReplayKey(tid)
	ctx := context.Background()

	for {
		val, err := actionRdb.Get(ctx, key).Result()
		if err == nil && val != "" {
			return val, nil
		}
		if err != nil && err != redis.Nil {
			logger.Log.Warn("luoping sync wait redis error",
				zap.String("tid", tid),
				zap.Error(err),
			)
		}
		if time.Now().After(deadline) {
			logger.Log.Warn("luoping sync wait timed out",
				zap.String("tid", tid),
				zap.Duration("timeout", timeout),
				zap.String("key", key),
			)
			return "", fmt.Errorf("等待设备回应超时")
		}
		time.Sleep(50 * time.Millisecond)
	}
}
