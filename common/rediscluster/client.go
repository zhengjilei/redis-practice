package rediscluster

import (
	"fmt"
	"github.com/go-redis/redis/v7"
)

var Client *redis.ClusterClient


func init() {
	Client = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{":30001", ":30002", ":30003", ":30004", ":30005", ":30006"},
	})
	result, err := Client.Ping().Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("ping result: ", result)
}
