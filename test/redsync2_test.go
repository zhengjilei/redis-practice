package tt

import (
	"github.com/go-redsync/redsync"
	"github.com/gomodule/redigo/redis"
	"os"
	"testing"
	"time"

	"github.com/stvp/tempredis"
)

var servers []*tempredis.Server

func TestMain(m *testing.M) {
	for i := 0; i < 8; i++ {
		server, err := tempredis.Start(tempredis.Config{})
		if err != nil {
			panic(err)
		}
		servers = append(servers, server)
	}
	result := m.Run()
	for _, server := range servers {
		server.Term()
	}
	os.Exit(result)
}

func TestRedsync(t *testing.T) {
	pools := newMockPools(8, servers)
	rs := redsync.New(pools)

	mutex := rs.NewMutex("test-redsync")
	err := mutex.Lock()
	if err != nil {
		
	}

}

func newMockPools(n int, servers []*tempredis.Server) []redsync.Pool {
	pools := []redsync.Pool{}
	for _, server := range servers {
		func(server *tempredis.Server) {
			pools = append(pools, &redis.Pool{
				MaxIdle:     3,
				IdleTimeout: 240 * time.Second,
				Dial: func() (redis.Conn, error) {
					return redis.Dial("unix", server.Socket())
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

