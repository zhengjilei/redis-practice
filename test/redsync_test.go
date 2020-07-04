package tt

import (
	"fmt"
	"github.com/go-redsync/redsync"
	"github.com/gomodule/redigo/redis"
	"testing"
	"time"
)

//redis命令执行函数
func DoRedisCmdByConn(conn *redis.Pool, commandName string, args ...interface{}) (interface{}, error) {
	redisConn := conn.Get()
	defer redisConn.Close()
	//检查与redis的连接
	return redisConn.Do(commandName, args...)
}

func TestRedis(t *testing.T) {

	//单个锁
	//pool := newPool()
	//rs := redsync.New([]redsync.Pool{pool})
	//mutex1 := rs.NewMutex("test-redsync1")
	//
	//mutex1.Lock()
	//conn := pool.Get()
	//conn.Do("SET","name1","ywb1")
	//conn.Close()
	//mutex1.Unxlock()
	curtime := time.Now().UnixNano()
	//多个同时访问
	pool := newPool()
	mutexes := newTestMutexes([]redsync.Pool{pool}, "test-mutex", 2)
	orderCh := make(chan int)
	for i, v := range mutexes {
		go func(i int, mutex *redsync.Mutex) {
			if err := mutex.Lock(); err != nil {
				t.Fatalf("Expected err == nil, got %q", err)
				return
			}
			fmt.Println(i, "add lock ....")
			conn := pool.Get()
			DoRedisCmdByConn(pool, "SET", fmt.Sprintf("name%v", i), fmt.Sprintf("name%v", i))
			str, _ := redis.String(DoRedisCmdByConn(pool, "GET", fmt.Sprintf("name%v", i)))
			fmt.Println(str)
			DoRedisCmdByConn(pool, "DEL", fmt.Sprintf("name%v", i))
			conn.Close()
			mutex.Unlock()
			fmt.Println(i, "del lock ....")
			orderCh <- i
		}(i, v)
	}
	for range mutexes {
		<-orderCh
	}
	fmt.Println(time.Now().UnixNano() - curtime)
}

func newTestMutexes(pools []redsync.Pool, name string, n int) []*redsync.Mutex {
	mutexes := []*redsync.Mutex{}
	for i := 0; i < n; i++ {
		mutexes = append(mutexes, redsync.New(pools).NewMutex(name,
			redsync.SetExpiry(time.Duration(2)*time.Second),
			redsync.SetRetryDelay(time.Duration(10)*time.Millisecond)),
		)
	}
	return mutexes
}

func newPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: time.Duration(24) * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "127.0.0.1:6379")
			if err != nil {
				panic(err.Error())
				//s.Log.Errorf("redis", "load redis redisServer err, %s", err.Error())
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			if err != nil {
				//s.Log.Errorf("redis", "ping redis redisServer err, %s", err.Error())
				return err
			}
			return err
		},
	}
}
