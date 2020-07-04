package semaphore_lua

import (
	"github.com/go-redis/redis/v7"
	"github.com/google/uuid"
	"github.com/zhengjilei/redis-practice/common/redisclient"
	"time"
)

const (
	timeout = 3000
)

// lua
var KEYS = []string{"", "semaphore:pet:timestamp", "semaphore:pet:owner", "semaphore:pet:count"}

var ARGS = []string{
	"",
	time.Now().UnixNano()/int64(time.Millisecond) - timeout, // timeoutThreshold
	"dsa4321fgvtrg", // identifier
	"143543654645",  // current time in milliseconds
	3,               // semaphore limit
}

// return 1: succeed to acquire semaphore 
// return 0: failed to acquire semaphore 

var acquireSemaphoreDesc = `
	

`
var acquireSemaphore = `
  redis.call("zremrangebyscore",KEYS[1],"-inf",ARGS[1])
  redis.call("zinterstore",KEYS[1],2, KEYS[1],KEYS[2],"WEIGHTS" ,0,1)
  local count = redis.call("incr",KEYS[3])
  
  redis.call("zadd",KEYS[2],count, ARGS[2])
  redis.call("zadd",KEYS[1],time, ARGS[3])
  local rank = redis.call("zrank",KEYS[2],ARGS[2])
  if rank < ARGS[4] then
       return 1
  else
      redis.call("zrem",KEYS[2],ARGS[2])
      redis.call("zrem",KEYS[1],ARGS[3])
      return 0
  end
`
var redisClient  = redisclient.Client
// KEYS[1] = "semaphore:pet:timestamp"
// ARGS[1] = current time in millisecond , ARGS[2] = identifier
var refreshSemaphore = `
	if redis.call("zadd",KEYS[1],ARGS[1],ARGS[2]) = 1
       redis.call("zrem", KEYS[1], ARGS[1])
       return false
    end
`

func AcquireSemaphore(timeout int64) string{
	//lockID:=acquireLock()
	//defer releaseLock(lockID)
	//
	//identifier:= uuid.New().String()
	//
	//while not timeout {
	//	if runAcquireScript(acquireSemaphore, KEYS, ARGS) == 1{
	//		return identifier
	//	}
	//}
	//return ""
}
