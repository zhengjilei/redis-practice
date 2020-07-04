package config

import (
	"errors"
	"fmt"
	"log"
	"github.com/zhengjilei/redis-practice/common/redisclient"
	"time"
)

const (
	configPrefix = "config"
)

var (
	checkedTimestamp = map[string]int64{}
	configCache      = map[string]string{}
)

func registerConfig(category, component string, jsonConfigInfo string) error {
	key := fmt.Sprintf("config:%s:%s", category, component)
	ok, err := redisclient.Client.Set(key, jsonConfigInfo, -1).Result()
	if err != nil {
		return err
	}
	if ok != "OK" {
		return errors.New("failed to register config")
	}
	return nil
}

// refreshTimeout 表示允许配置的缓存时间（秒）
func getConfig(category, component string, refreshTimeout int64) (string, error) {
	key := fmt.Sprintf("config:%s:%s", category, component)
	ts, ok := checkedTimestamp[key]
	if ok && ts > (time.Now().Unix()-refreshTimeout) {
		return configCache[key], nil
	}
	// 没有缓存，或者缓存已超时
	log.Println("no cache or cache timeout")

	val, err := redisclient.Client.Get(key).Result()
	if err != nil {
		return "", err
	}
	configCache[key] = val
	checkedTimestamp[key] = time.Now().Unix()
	return val, nil
}
