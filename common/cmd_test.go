package common

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/zhengjilei/redis-practice/common/redisclient"
	"testing"
	"time"
)

func TestGetErr(t *testing.T) {

	s, err := redisclient.Client.Get("aaaa").Result()
	fmt.Println(err)              // redis: nil
	fmt.Println(s)                // ""
	fmt.Println(err == redis.Nil) // true 

	s2 := redisclient.Client.Get("bbbbbbb").String()
	fmt.Println(s2) // get bbbbbbb: redis: nil

}
func TestIsMember(t *testing.T) {
	flag, err := redisclient.Client.SIsMember("aa", "vvv").Result()
	fmt.Println(err)              // <nil>
	fmt.Println(err == redis.Nil) // false
	fmt.Println(flag)             // false

	flag2 := redisclient.Client.SIsMember("aa", "vvv").Val()
	fmt.Println(flag2) // false

	// wrong type
	flag3, err := redisclient.Client.SIsMember("mzset", "vvv").Result()
	fmt.Println(err)              // WRONGTYPE Operation against a key holding the wrong kind of value
	fmt.Println(err == redis.Nil) // false
	fmt.Println(flag3)            // false

	flag4 := redisclient.Client.SIsMember("aa", "vvv").Val()
	fmt.Println(flag4) // false

}

func TestPipeline(t *testing.T) {
	var incr *redis.IntCmd
	_, err := redisclient.Client.Pipelined(func(pipe redis.Pipeliner) error {
		incr = pipe.Incr("pipelined_counter")
		pipe.Expire("pipelined_counter", time.Hour)
		return nil
	})
	fmt.Println(incr.Val(), err)
}
