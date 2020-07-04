package main

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/google/uuid"
	"github.com/zhengjilei/redis-practice/common/redisclient"
	"time"
)

func main2() {
	err := redisclient.Client.Set("mkey", "v1", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := redisclient.Client.Get("mkey").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(val)

	val2, err := redisclient.Client.Get("mkey2").Result()
	if err == redis.Nil {
		fmt.Println("mkey2 does not exist")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println(val2)
	}
}
func main() {
	//now := time.Now().UTC().Unix()
	//fmt.Println(now)
	fmt.Println(uuid.New().String())
	fmt.Println(uuid.New().String())
	fmt.Println(uuid.New().String())

	fmt.Println(time.Now().UTC().UnixNano())
	fmt.Println(time.Now().UTC().Unix())
	fmt.Println(time.Now().Local().Unix())
	
	<- time.After(3*time.Second)
	fmt.Println("main over")
}
