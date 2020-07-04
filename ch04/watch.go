package main

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"log"
	"github.com/zhengjilei/redis-practice/common/rediscluster"
	"sync"
)

func main() {

	ExampleClient_Watch()
}
func ExampleClient_Watch() {

	rdb := rediscluster.Client

	const routineCount = 3

	// Transactionally increments key using GET and SET commands.
	increment := func(key string) error {
		txf := func(tx *redis.Tx) error {
			// get current value or zero
			n, err := tx.Get(key).Int()
			if err != nil && err != redis.Nil {
				return err
			}

			// actual opperation (local in optimistic lock)
			n++

			// runs only if the watched keys remain unchanged

			_, err = tx.Pipelined(func(pipe redis.Pipeliner) error {
				// pipe handles the error case
				pipe.Set(key, n, -1)
				return nil
			})
			return err
		}

		for retries := routineCount; retries > 0; retries-- {
			err := rdb.Watch(txf, key)
			if err == nil {
				return nil
			}
			if err == redis.TxFailedErr {
				log.Println("watch failed, retry: ", 100-retries)
				return err
			}
			log.Printf("not watch err:[%+v] retry:%d\n", err, 100-retries)
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
				fmt.Println("increment error:", err)
			}
		}()
	}
	wg.Wait()

	n, err := rdb.Get("counter3").Int()
	fmt.Println("ended with", n, err)
	// Output: ended with 100 <nil>
}
