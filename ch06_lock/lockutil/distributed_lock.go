package lockutil

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/google/uuid"
	"time"
)

const (
	defaultLockTimeout = 5
)

func getLockKey(lockName string) string {
	return fmt.Sprintf("lock:%s", lockName)
}

// 未给锁设置超时时间：可能会有些客户端获取锁之后崩溃了，其他客户端一直在等待
func AcquireLock(client *redis.Client, lockName string, timeoutInSeconds int) (string, error) {
	identifier := uuid.New().String()
	end := time.Now().Add(time.Duration(timeoutInSeconds) * time.Second)
	for ; time.Now().Before(end); {
		// 实际上调用的 set key value ex -1 nx
		if client.SetNX(getLockKey(lockName), identifier, -1).Val() {
			return identifier, nil
		}
		// 减少重试的频率
		time.Sleep(time.Millisecond * 10)
	}
	return "", errors.New("failed to get lock with timeout")
}

// 在2.6.12 之前，没有 set key value ex timeout nx 原子命令，需要执行两步
// 1. setnx key value
// 2. expire key timeout
// 需要在获取锁之后，设置锁的超时时间，超时之后锁自动释放
// 可能在获取锁之后客户端崩溃，所以其他client 在获取不到锁时需要检测锁是否有超时时间，没有需要设置
func AcquireLockV2(client *redis.Client, lockName string, retryTimeout, lockReleaseTimeout int) (string, error) {
	identifier := uuid.New().String()
	end := time.Now().Add(time.Duration(retryTimeout) * time.Second)
	lockKey := getLockKey(lockName)
	for ; time.Now().Before(end); {
		// 实际上调用的 set key value ex 10 nx, 已经设置超时时间了
		if client.SetNX(lockKey, identifier, time.Duration(lockReleaseTimeout)).Val() {
			return identifier, nil
		} else {
			// 防止有些锁没有设置超时时间
			ttl := client.TTL(lockKey).Val()
			if ttl == -1 {
				client.Expire(lockKey, time.Duration(lockReleaseTimeout))
			}
		}
		time.Sleep(time.Millisecond * 10)
	}
	return "", errors.New("failed to get lock with timeout")
}

// Final version: 只适用于单机
func AcquireLockV3(client *redis.Client, lockName string, retryTimeout, lockReleaseTimeout int) (string, error) {
	identifier := uuid.New().String()
	end := time.Now().Add(time.Duration(retryTimeout) * time.Second)
	lockKey := getLockKey(lockName)
	for ; time.Now().Before(end); {
		if client.SetNX(lockKey, identifier, time.Duration(lockReleaseTimeout)*time.Second).Val() {
			return identifier, nil
		}
		time.Sleep(time.Millisecond * 10)
	}
	return "", errors.New("failed to get lock with timeout")
}

const (
	acquireLockLua = ` if redis.call("set", KEYS[1],ARGV[1],"ex",ARGV[2],"nx") then 
                    return 1 
                   else 
                    return 0
                   end`

	releaseLockLua = `if redis.call("get",KEYS[1]) == ARGV[1] then
             		     redis.call("del",KEYS[1])
                  		 return 1
					  else
						  return 0
					  end`
)

// https://www.zhihu.com/question/300767410?sort=created
//1.2 释放锁
// lua 脚本， 原子，不需要 watch 命令
//if redis.call("get",KEYS[1]) == ARGV[1] then
//    return redis.call("del",KEYS[1])
//else
//    return 0
//end
func ReleaseLock(client *redis.Client, lockName, identifier string) error {
	lockKey := getLockKey(lockName)

	ftx := func(tx *redis.Tx) error {
		res, err := tx.Get(lockKey).Result()
		if err == nil && res == identifier {
			// 满足释放锁的条件
			_, err = tx.Pipelined(func(pipe redis.Pipeliner) error {
				tx.Del(lockKey)
				return nil
			})
			// 可能为 Nil
			return err
		}

		// 获取失败
		if err != nil {
			return err
		}
		if res != identifier {
			return fmt.Errorf("release lock failed, lock: %s was changed to %s", identifier, res)
		}
		return err
	}

	for ; ; {
		if err := client.Watch(ftx, lockKey); err != redis.TxFailedErr {
			// 非 watch err, 直接返回
			return err
		}
		//冲突，再次尝试
		time.Sleep(time.Millisecond * 10)
	}
}
