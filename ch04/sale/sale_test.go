package main

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/zhengjilei/redis-practice/common/rediscluster"
	"sync"
	"testing"
	"time"
)

func TestRedisCluster(t *testing.T) {
	client := rediscluster.Client
	res, err := client.SMembers("inventory:27").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(res)

}

func TestWatchCrossSlot(t *testing.T) {
	rdb := rediscluster.Client
	defer rdb.Close()
	err := rdb.Watch(func(tx *redis.Tx) error {
		fmt.Println("hjh")
		return nil
	}, "a", "bc", "market")
	fmt.Println(err)
}

// Why all success？
func TestCrossSlot(t *testing.T) {
	rdb := rediscluster.Client
	defer rdb.Close()
	tx := rdb.TxPipeline()
	tx.Set("abcde:1234", 999, 100*time.Second)
	tx.Set("abcde:9876", 1000, 200*time.Second)
	tx.HSet("user-profile:1234", "username", "king foo")
	tx.HSet("user-session:1234", "username", "king foo")
	r, err := tx.MSet("a", 123, "bbbc", 456).Result()
	fmt.Println(err)
	fmt.Println(r)
	fmt.Println("-----------")
	res, err := tx.Exec()
	if err != nil {
		fmt.Println(err)
	}
	for _, v := range res {
		fmt.Println(v.Name(), v.Args(), v.Err())
	}
}

func TestCrossSlot2(t *testing.T) {
	rediscluster.Client.MSet("a", 123, "bbbc", 456)
}
func TestZadd(t *testing.T) {
	rdb := rediscluster.Client

	res, err := rdb.TxPipelined(func(pipe redis.Pipeliner) error {
		pipe.ZAdd("market", &redis.Z{
			Score:  21,
			Member: "abc",
		})
		pipe.SAdd("inventory:27", "ItemA").Result()
		return nil
	})
	fmt.Println(err)
	for _, v := range res {
		fmt.Println(v.Name(), v.Args(), v.Err())
	}
	rdb.Close()
}

var (
	retryCount = 0
)

func TestWatchCluster(t *testing.T) {
	rdb := rediscluster.Client

	const routineCount = 100

	var mutex sync.Mutex

	// Transactionally increments key using GET and SET commands.
	increment := func(key string) error {
		txf := func(tx *redis.Tx) error {
			//fmt.Println("start txf func")
			// get current value or zero
			n, err := tx.Get(key).Int()
			if err != nil && err != redis.Nil {
				return err
			}

			// actual opperation (local in optimistic lock)
			n++

			// runs only if the watched keys remain unchanged
			result, err := tx.Pipelined(func(pipe redis.Pipeliner) error {
				// pipe handles the error case
				pipe.Set(key, n, 0)
				//pipe.ZAdd(marketKey, &redis.Z{
				//	Score:  21,
				//	Member: "abc",
				//})
				//pipe.SAdd(inventoryKey, "ItemA")
				return nil
			})
			_ = result
			//util.PrintResult(result)
			return err
		}

		for retries := routineCount; retries > 0; retries-- {
			//err := rdb.Watch(txf, key, marketKey, "inventory:27")
			err := rdb.Watch(txf, key)
			// 没有执行 txf，直接报错，因为watch 检测到多个 key 不在同一个 slot 
			if err != redis.TxFailedErr {
				return err
			}
			mutex.Lock()
			retryCount++
			mutex.Unlock()
			//fmt.Println(err)
			// optimistic lock lost
		}
		return errors.New("increment reached maximum number of retries")
	}

	var wg sync.WaitGroup
	wg.Add(routineCount)
	for i := 0; i < routineCount; i++ {
		go func() {
			defer wg.Done()
			if err := increment("counter3"); err != nil {
				// redis: Watch requires all keys to be in the same slot
				fmt.Println("increment error:", err)
			}
		}()
	}
	wg.Wait()

	n, err := rdb.Get("counter3").Int()
	fmt.Println("retryCount: ", retryCount)
	fmt.Println("ended with", n, err)

}
