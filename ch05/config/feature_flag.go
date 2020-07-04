package config

import (
	"errors"
	"github.com/zhengjilei/redis-practice/common/redisclient"
	"strconv"
	"time"
)

const (
	maintenanceFeatureFlagKey = "is-under-maintenance"
)

var (
	lastCheckedTimeStamp  int64
	isUnderMaintenanceVal = false
)

func isUnderMaintenance() bool {
	if lastCheckedTimeStamp < time.Now().Unix()-1 {
		lastCheckedTimeStamp = time.Now().Unix()
		isUnderMaintenanceVal = flag(maintenanceFeatureFlagKey, false)
	}
	return isUnderMaintenanceVal
}

func registerFlag(key string, val bool) error {
	ok, err := redisclient.Client.Set(key, strconv.FormatBool(val), -1).Result()
	if err != nil {
		return err
	}
	if ok != "OK" {
		return errors.New("failed to register feature flag")
	}
	return nil
}

// df default value
func flag(key string, df bool) bool {
	ok, err := redisclient.Client.Get(key).Result()
	if err != nil {
		return df
	}

	switch ok {
	case "true":
		return true
	case "false":
		return false
	default:
		return df
	}
}
