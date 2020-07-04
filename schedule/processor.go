package main

import (
	"github.com/go-redis/redis/v7"
	"github.com/zhengjilei/redis-practice/common/redisclient"
	"time"
)

const (
	petTaskZSetKey = "pet_task"
)

func main() {
}

func saveTask(taskID int64) error {
	_, err := redisclient.Client.ZAdd(petTaskZSetKey, &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: taskID,
	}).Result()
	return err
}

// retrieve task and delete it in transaction
func retrieveTask() int64 {
	redisclient.Client.ZRange(petTaskZSetKey, 0, 0).Result()
}
