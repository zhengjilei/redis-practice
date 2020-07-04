package test

import (
	"fmt"
	"github.com/zhengjilei/redis-practice/common/redisclient"
	"sync"
	"testing"
	"time"
)

const (
	countKey = "count"
)

func TestTran(t *testing.T) {
	go func() {
		pipe := redisclient.Client.TxPipeline()
		defer pipe.Close()
		incr1 := pipe.Incr(countKey)
		fmt.Println("incr1 over")
		time.Sleep(10 * time.Second)
		incr2 := pipe.Incr(countKey)
		_, err := pipe.Exec()
		if err != nil {
			panic(err)
		}
		fmt.Println("incr1 :", incr1.Val())
		fmt.Println("incr2 :", incr2.Val())
	}()

	time.Sleep(1 * time.Second)
	v, err := redisclient.Client.Get(countKey).Int64()
	if err != nil {
		panic(err)
	}
	fmt.Println("count: ", v)
	time.Sleep(5 * time.Second)
	redisclient.Client.Set(countKey, 1000, -1)
	fmt.Println("main set 1000 over")
	time.Sleep(10 * time.Second)

	v, err = redisclient.Client.Get(countKey).Int64()
	if err != nil {
		panic(err)
	}
	fmt.Println("end count: ", v)
}

func TestTranErr(t *testing.T) {
	v, err := redisclient.Client.Get(countKey).Int64()
	if err != nil {
		panic(err)
	}
	fmt.Println("count: ", v)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		tx := redisclient.Client.TxPipeline()
		defer tx.Close()
		tx.Incr(countKey)
		tx.Incr("non-exist-key")
		tx.Incr(countKey)
		resp, err := tx.Exec()
		if err != nil {
			fmt.Println(err)
		}
		if resp != nil && len(resp) > 0 {
			for _, v := range resp {
				fmt.Println(v.Name(), v.Err(), v.Args())
			}
		}
		wg.Done()
	}()
	wg.Wait()
	v, err = redisclient.Client.Get(countKey).Int64()
	if err != nil {
		panic(err)
	}
	fmt.Println("end count: ", v)
}
