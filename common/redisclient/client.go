package redisclient

import (
	"fmt"

	"github.com/go-redis/redis/v7"
)

var Client *redis.Client

func init() {
	Client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	pong, err := Client.Ping().Result()
	fmt.Println(pong, err)
}
