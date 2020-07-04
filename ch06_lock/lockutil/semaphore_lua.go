package lockutil

import (
	"github.com/go-redis/redis/v7"
	"github.com/zhengjilei/redis-practice/common/redisclient"
	"time"
)

const (
	timeout = 3000
)

var redisClient = redisclient.Client

// lua
var KEYS = []string{"", "semaphore:pet:timestamp", "semaphore:pet:owner", "semaphore:pet:count"}

// return 1: succeed to acquire semaphore 
// return 0: failed to acquire semaphore 

var acquireSemaphoreDesc = `
	redis.call("zremrangebyscore","semaphore:remote","-inf",time.Now()-timeout)
	redis.call("zinterstore","semaphore:remote:owner",2,"semaphore:remote","semaphore:remote:owner","WEIGHTS",0,1)
	local cnt  = redis.call("incr","semaphore:remote:counter")
	if redis.call("zadd","semaphore:remote:owner",cnt,identifier) == 1 then 
		if redis.call("zrank","semaphore:remote:owner",identifier) < limit then 
			redis.call("zadd","semaphore:remote",time.Now(), identifier)
			return true
		else
			redis.call("zrem","semaphore:remote:owner",identifier)
			return false
		end
	else
		return false
	end
`

// KEYS[1] = semaphore:-:timestamp, KEYS[2] = semaphore:-:owner, KEYS[3] = semaphore:-:counter
// ARGV[1] = timeout_threshold, ARGV[2]= identifier,ARGV[3] = limit, ARGV[4] = time.Now()
var acquireSemaphore = `
  redis.call("zremrangebyscore",KEYS[1],"-inf",ARGV[1])
  redis.call("zinterstore",KEYS[2],2, KEYS[1],KEYS[2],"WEIGHTS" ,0,1)
  local cnt = redis.call("incr",KEYS[3])
  if redis.call("zadd",KEYS[2],cnt,ARGV[2]) == 1 then 
  	if redis.call("zrank",KEYS[2],ARGV[2]) < ARGV[3] then 
  		redis.call("zadd",KEYS[1],ARGV[4], ARGV[2])
  		return true
  	else
  		redis.call("zrem",ARGV[2],ARGV[2])
  		return false
  	end
  else
  	return false
  end
`

var releaseSemaphoreDesc = `
	redis.call("zrem","semaphore:remote",identifier)
	redis.call("zrem","semaphore:remote:owner",identifier)
`
var releaseSemaphore = `
	redis.call("zrem",KEYS[1],ARGV[1])
	redis.call("zrem",KEYS[2],ARGV[1])
`

// 0 表示 刷新失败，信号量已经超时删除了
var refreshSemaphoreDesc = `
	if redis.call("zadd","semaphore:remote",time.Now(),identifier) == 1 then
		redis.call("zrem","semaphore:remote",identifier)
		redis.call("zrem","semaphore:remote:owner",identifier)
		return false
	else 
		return true
`

// KEYS[1] = "semaphore:remote", KEYS[2] = "semaphore:remote:owner"
// ARGV[1] = current time in nanoseconds , ARGV[2] = identifier
var refreshSemaphore = `V
	if redis.call("zadd",KEYS[1],ARGV[1],ARGV[2]) == 1
        redis.call("zrem", KEYS[1], ARGV[1])
		redis.call("zrem",KEYS[2],identifier)
        return false
	else 
		return true
    end
`

var (
	acquireSemaphoreScript = redis.NewScript(acquireSemaphore)
	releaseSemaphoreScript = redis.NewScript(releaseSemaphore)
	refreshSemaphoreScript = redis.NewScript(refreshSemaphore)
)

type SemaphoreLua struct {
	timestampKey string
	ownerKey     string // semaphore owner
	counterKey   string
	timeout      int64 // in seconds, timeout 指的是信号量本身的超时时间,不是获取不到信号量重试的超时时间
	limit        int   // the limit of process at one time
	client       redis.Client
}

func NewSemaphoreLua(client redis.Client, taskIdentifier string, timeout int64, limit int) *SemaphoreLua {
	return &SemaphoreLua{
		client:       client,
		timestampKey: "semaphore:" + taskIdentifier + ":timestamp",
		ownerKey:     "semaphore:" + taskIdentifier + ":owner",
		counterKey:   "semaphore:" + taskIdentifier + ":count",
		timeout:      timeout,
		limit:        limit,
	}
}

func (s *SemaphoreLua) AcquireSemaphore(identifier string) (bool, error) {
	now := time.Now().UnixNano()
	ok, err := acquireSemaphoreScript.Run(s.client, []string{s.timestampKey, s.ownerKey, s.counterKey}, now-s.timeout*int64(time.Second), identifier, s.limit, now).Bool()
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (s *SemaphoreLua) ReleaseSemaphore(identifier string) error {
	return releaseSemaphoreScript.Run(s.client, []string{s.timestampKey, s.ownerKey}, identifier).Err()
}

func (s *SemaphoreLua) RefreshSemaphore(identifier string) (bool, error) {
	ok, err := refreshSemaphoreScript.Run(s.client, []string{s.timestampKey, s.ownerKey}, time.Now().UnixNano(), identifier).Bool()
	if err != nil {
		return false, nil
	}
	return ok, nil
}
