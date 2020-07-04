package lockutil

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/google/uuid"
	"strconv"
	"time"
)

var (
	ErrLostSemaphore      = errors.New("refresh failed because the semaphore was already expired")
	ErrGetSemaphoreFailed = errors.New("get semaphore failed, maybe need retry")
)

func getOwnerKey(semname string) string {
	return fmt.Sprintf("%s:owner", semname)
}
func getCounterKey(semname string) string {
	return fmt.Sprintf("%s:counter", semname)
}

// 解决上面系统时间误差的问题，但仍然存在多客户端进程导致信号量的持有者比预期的多的问题
// P129
func AcquireFairSemaphore(client *redis.Client, semname string, limit int, timeoutInSecond int64) (string, error) {
	nowInMilli := time.Now().UnixNano() / int64(time.Millisecond)
	ownerKey := getOwnerKey(semname)
	counterKey := getCounterKey(semname)
	identifier := uuid.New().String()

	var countResp *redis.IntCmd
	_, err := client.Pipelined(func(pipe redis.Pipeliner) error {
		// 1. 从超时信号量集合中移除超时的信号量
		pipe.ZRemRangeByScore(semname, "-inf", strconv.Itoa(int(nowInMilli-timeoutInSecond*int64(time.Millisecond))))
		// 2. 超时信号量集合和 owner 集合作交集  
		pipe.ZInterStore(ownerKey, &redis.ZStore{
			Keys:    []string{ownerKey, semname},
			Weights: []float64{1, 0},
		})
		// 3. 递增 counter，取得 counter 的最新值 
		countResp = pipe.Incr(counterKey)
		return nil
	})
	if err != nil {
		return "", err
	}
	if countResp == nil || countResp.Err() != nil {
		return "", errors.New("get error when incrementing counter key")
	}

	var rank *redis.IntCmd
	_, err = client.Pipelined(func(pipe redis.Pipeliner) error {
		// 4. 保存 counter 到 owner 集合中
		pipe.ZAdd(counterKey, &redis.Z{
			Score:  float64(countResp.Val()),
			Member: identifier,
		})
		// 5. 添加当前的时间到超时信号量集合中
		pipe.ZAdd(semname, &redis.Z{
			Score:  float64(nowInMilli),
			Member: identifier,
		})
		// 6. 得到当前客户端新添加的值在 owner 集合中的排名
		rank = pipe.ZRank(ownerKey, identifier)
		return nil
	})
	if err != nil {
		return "", err
	}
	if rank == nil || rank.Err() != nil {
		return "", errors.New("get error when executing zrank")
	}
	// 排名小于 limit则说明获得信号量，直接返回 identifier
	if int(rank.Val()) < limit {
		return identifier, nil
	}
	// 大于等于 limit,没有获得信号量，从 超时信号量集合和 owner 集合中移除新添加的
	_, _ = client.Pipelined(func(pipe redis.Pipeliner) error {
		pipe.ZRem(semname, identifier)
		pipe.ZRem(counterKey, identifier)
		return nil
	})
	return "", ErrGetSemaphoreFailed
}

func ReleaseFairSemaphore(client *redis.Client, semname, identifier string) bool {
	_, err := client.Pipelined(func(pipe redis.Pipeliner) error {
		pipe.ZRem(semname, identifier)
		pipe.ZRem(getOwnerKey(semname), identifier)
		return nil
	})
	return err == nil
}

func RefreshFairSemaphore(client *redis.Client, semname, identifier string) (bool, error) {
	addCount, err := client.ZAdd(semname, &redis.Z{
		Score:  float64(time.Now().UnixNano() / int64(time.Millisecond)),
		Member: identifier,
	}).Result()
	if err != nil {
		return false, err
	}
	if addCount == 1 {
		// 说明该信号量是新加的，之前的信号量已经超时移除，所以不能 refresh
		client.ZRem(semname, identifier)
		return false, ErrLostSemaphore
	}

	// addCount == 0, refresh 成功
	return true, nil
}

// timeoutInSecond 表示信号量的超时时间
func AcquireFairSemaphoreWithLock(client *redis.Client, semname string, limit int, timeoutInSecond int64) (string, error) {
	identifier, err := AcquireLockV3(client, semname, 10, 10)
	if err != nil {
		return "", err
	}
	defer ReleaseLock(client, semname, identifier)

	nowInMilli := time.Now().UnixNano() / int64(time.Millisecond)
	ownerKey := getOwnerKey(semname)
	counterKey := getCounterKey(semname)

	var countResp *redis.IntCmd
	_, err = client.Pipelined(func(pipe redis.Pipeliner) error {
		// 1. 从超时信号量集合中移除超时的信号量
		pipe.ZRemRangeByScore(semname, "-inf", strconv.Itoa(int(nowInMilli-timeoutInSecond*int64(time.Millisecond))))
		// 2. 超时信号量集合和 owner 集合作交集  
		pipe.ZInterStore(ownerKey, &redis.ZStore{
			Keys:    []string{ownerKey, semname},
			Weights: []float64{1, 0},
		})
		// 3. 递增 counter，取得 counter 的最新值 
		countResp = pipe.Incr(counterKey)
		return nil
	})
	if err != nil {
		return "", err
	}
	if countResp == nil || countResp.Err() != nil {
		return "", errors.New("get error when incrementing counter key")
	}

	var rank *redis.IntCmd
	_, err = client.Pipelined(func(pipe redis.Pipeliner) error {
		// 4. 保存 counter 到 owner 集合中
		pipe.ZAdd(ownerKey, &redis.Z{
			Score:  float64(countResp.Val()),
			Member: identifier,
		})
		// 5. 添加当前的时间到超时信号量集合中
  		// 6. 得到当前客户端新添加的值在 owner 集合中的排名
		rank = pipe.ZRank(ownerKey, identifier)
		return nil
	})
	if err != nil {
		return "", err
	}
	if rank == nil || rank.Err() != nil {
		return "", errors.New("get error when executing zrank")
	}
	// 排名小于 limit则说明获得信号量，直接返回 identifier
	if int(rank.Val()) < limit {
		return identifier, nil
	}
	// 大于等于 limit,没有获得信号量，从 超时信号量集合和 owner 集合中移除新添加的
	_, _ = client.Pipelined(func(pipe redis.Pipeliner) error {
		pipe.ZRem(semname, identifier)
		pipe.ZRem(counterKey, identifier)
		return nil
	})
	return "", ErrGetSemaphoreFailed
}
