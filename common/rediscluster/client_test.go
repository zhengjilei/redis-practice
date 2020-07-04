package rediscluster

import (
	"fmt"
	"log"
	"testing"
)

func TestRedis(t *testing.T) {
	Client.Set("a", 10, -1)
	a, err := Client.Get("a").Int64()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(a)
}
