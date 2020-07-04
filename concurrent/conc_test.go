package concurrent

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/zhengjilei/redis-practice/common/redisclient"
	"sync"
	"testing"
)

func TestConcurrentPop(t *testing.T) {

	client := redisclient.Client
	defer client.Close()

	concList := "conlist"

	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(k int) {
			defer wg.Done()
			for j := k * 100; j < k*100+100; j++ {
				client.LPush(concList, j)
			}
		}(i)
	}
	wg.Wait()

	count := client.LLen(concList).Val()
	fmt.Println("count: ", count) // 10000 

	consumeCount := 0

	lock := sync.Mutex{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ; ; {
				if err := client.RPop(concList).Err(); err == redis.Nil {
					break
				}
				lock.Lock()
				consumeCount++
				lock.Unlock()
			}
		}()
	}
	wg.Wait()
	fmt.Println("consume count: ", consumeCount)
}

func TestPop(t *testing.T) {
	str, err := redisclient.Client.RPop("mlist").Result()
	fmt.Println(str)
	fmt.Println(err)
	fmt.Println(err == redis.Nil)
}
