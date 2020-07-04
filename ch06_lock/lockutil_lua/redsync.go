package lockutils

import (
	"github.com/go-redsync/redsync"
	"github.com/gomodule/redigo/redis"
	"github.com/stvp/tempredis"
	"time"
)

var sync *redsync.Redsync

func createPool(url string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 180 * time.Second, // Default is 300 seconds for redis server
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", url)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func Initialize(url string) {
	pools := []redsync.Pool{
		createPool(url),
	}
	sync = redsync.New(pools)
}


func AcquireLock() {
	
}
func newPools(n int, servers []*tempredis.Server) []redsync.Pool {
	pools := []redsync.Pool{}
	for _, server := range servers {
		func(server *tempredis.Server) {
			pools = append(pools, &redis.Pool{
				MaxIdle:     3,
				IdleTimeout: 240 * time.Second,
				Dial: func() (redis.Conn, error) {
					return redis.Dial("tcp", server.Socket())
				},
				TestOnBorrow: func(c redis.Conn, t time.Time) error {
					_, err := c.Do("PING")
					return err
				},
			})
		}(server)
		if len(pools) == n {
			break
		}
	}
	return pools
}

//func AcquireLock(timeoutInSeconds int) (bool, error) {
//	end := time.Now().Add(time.Duration(timeoutInSeconds) * time.Second)
//	for ; time.Now().Before(end); {
//		redsync.Mutex{}
//	}
//
//}
