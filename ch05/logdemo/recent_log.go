package main

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/zhengjilei/redis-practice/common/util"
	"time"
)

const (
	logDebug    = "logDebug"
	logInfo     = "logInfo"
	logWarn     = "logWarn"
	logError    = "logError"
	logCritical = "logCritical"
)

var (
	logLevelMap = map[string]string{
		logDebug:    "debug",
		logInfo:     "info",
		logWarn:     "warn",
		logError:    "error",
		logCritical: "critical",
	}
)
// list   recent:{name}:info
func getRecentKey(name, logLevel string) string {
	return fmt.Sprintf("recent:%s:%s", name, logLevel)
}
func logRecent(pipe redis.Pipeliner, name, msg, logType string) error {
	logLevel, ok := logLevelMap[logType]
	if !ok {
		return errors.New("invalid log level")
	}
	destination := getRecentKey(name, logLevel)
	message := time.Now().String() + " " + msg
		
	pipe.LPush(destination, message)
	pipe.LTrim(destination, 0, 99) // 只保留最左边的100条信息
	result, err := pipe.Exec()
	if err != nil {
		return err
	}
	util.PrintResult(result)
	return nil
}
