package main

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/zhengjilei/redis-practice/common/rediscluster"
)

func main() {
	str := rediscluster.Client.Info("master_link_status").String()
	fmt.Println(str)
	rediscluster.Client.ForEachMaster(func(client *redis.Client) error {
		//
		return nil
	})

}
