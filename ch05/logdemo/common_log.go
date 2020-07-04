package main

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"strconv"
	"time"
)

func getCommonKey(name, logLevel string) string {
	return fmt.Sprintf("common:%s:%s", name, logLevel)
}

// zset  common:{name}:info      member 是 消息，score 是计数
// string common:{name}:info:start  存储当前的小时   23

// zset common:{name}:info:last
// string common:{name}:info:start:pstart
func logCommon(name, msg, logType string) error {
	logLevel, ok := logLevelMap[logType]
	if !ok {
		return errors.New("invalid log level")
	}
	destination := getCommonKey(name, logLevel) // zset
	startKey := destination + ":start"          // 记录最新的小时数
	end := time.Now().Add(5 * time.Second)

	fnx := func(tx *redis.Tx) error {
		nowHour := time.Now().Hour()
		s, err := tx.Get(startKey).Result()
		if err != redis.Nil {
			return err
		}
		getHour, _ := strconv.ParseInt(s, 10, 64)

		pip := tx.Pipeline()
		if getHour != 0 && int(getHour) < nowHour {
			pip.Rename(destination, destination+":last")
			pip.Rename(startKey, destination+":pstart")
			pip.Set(startKey, nowHour, -1)
		} else if getHour == 0 {
			pip.Set(startKey, nowHour, -1)
		}
		pip.ZIncrBy(destination, 1, msg)
		return logRecent(pip, name, msg, logType)
	}
	for ; time.Now().Before(end); {
		if err := client.Watch(fnx, startKey); err != redis.TxFailedErr {
			return err
		}
	}
	return nil
}
