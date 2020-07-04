package main

import (
	"github.com/go-redis/redis/v7"
	"github.com/zhengjilei/redis-practice/common/redisclient"
	"time"
)

var client *redis.Client

func main() {
	//client := rediscluster.Client
	client = redisclient.Client
	defer client.Close()

	//prepareData()
	//err := putIntoMarket("ItemM", 17, 97)
	//fmt.Println("main: ", err)

	//clearData()
	//purchaseItem(27, 17, "ItemM")
}
func clearData() {
	client.Expire("market", 0*time.Second)
	client.Expire("inventory:17",0*time.Second)
	client.Expire("inventory:27",0*time.Second)
	client.Expire("users:17",0*time.Second)
	client.Expire("users:27",0*time.Second)
}
func prepareData() {
	_, err := client.Pipelined(func(pipe redis.Pipeliner) error {
		pipe.HMSet("users:17", "name", "Frank", "funds", 43)
		pipe.SAdd("inventory:17", "ItemL", "ItemM", "ItemN")
		pipe.HMSet("users:27", "name", "Bill", "funds", 125)
		pipe.SAdd("inventory:27", "ItemO", "ItemP", "ItemQ")
		pipe.ZAdd("market", &redis.Z{
			Score:  35,
			Member: "ItemA.4",
		}, &redis.Z{
			Score:  48,
			Member: "ItemC.7",
		}, &redis.Z{
			Score:  60,
			Member: "ItemE.2",
		}, &redis.Z{
			Score:  73,
			Member: "ItemG.3",
		})
		return nil
	})
	if err != nil {
		panic(err)
	}

}
