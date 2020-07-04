package lockutil

import (
	"errors"
	"github.com/go-redis/redis/v7"
	"github.com/google/uuid"
	"strconv"
	"time"
)

// 大多数情况下没有问题，且运行速度快
// 当多个系统同时使用，部分系统之间时间出现误差时，会出现某个系统偷走另外一个系统的信号量
func AcquireSemaphore(client *redis.Client, semname string, limit int, timeoutInSecond int64) (string, error) {
	identifier := uuid.New().String()
	nowInMilli := time.Now().UnixNano() / int64(time.Millisecond)

	result, err := client.Pipelined(func(pipe redis.Pipeliner) error {
		// 移除掉超时的占用信号量的进程（客户端）
		pipe.ZRemRangeByScore(semname, "-inf", strconv.Itoa(int(nowInMilli-timeoutInSecond*int64(time.Millisecond))))
		// 将当前客户端加入信号量集合
		pipe.ZAdd(semname, &redis.Z{
			Score:  float64(nowInMilli),
			Member: identifier,
		})
		// 获取刚刚加入值的排名
		pipe.ZRank(semname, identifier)
		return nil
	})
	if err != nil {
		return "", err
	}
	if len(result) != 3 {
		return "", errors.New("invalid length of result")
	}

	rank, ok := result[2].(*redis.IntCmd)
	if !ok {
		return "", errors.New("invalid return type")
	}
	// 排名在前 Limit 名的话才说明获取信号量成功
	if rank.Val() < int64(limit) {
		// 获得信号量成功
		return identifier, nil
	}

	// 获取信号量失败，需要从 zset 中删除插入的值
	client.ZRem(semname, identifier)
	return "", errors.New("failed to get semaphore")
}

// true: 释放信号量成功
// false: 信号量已经超时，释放信号量出错
func ReleaseSemaphore(client *redis.Client, semname, identifier string) bool {
	return client.ZRem(semname, identifier).Val() > 0
}


