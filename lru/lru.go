package main

import (
	"github.com/go-redis/redis/v7"
	"github.com/zhengjilei/redis-practice/common/redisclient"
)

const (
	listKey       = "mlist"
	limit         = 100
	maxRetryCount = 50
)

var (
	client = redisclient.Client
)

func Put(val string) error {
	ftx := func(tx *redis.Tx) error {
		_, err := tx.Pipelined(func(pipe redis.Pipeliner) error {
			pipe.LRem(listKey, 1, val)
			pipe.LPush(listKey, val)
			pipe.LTrim(listKey, 0, 99)
			return nil
		})
		return err
	}

	for i := 0; i < maxRetryCount; i++ {
		err := client.Watch(ftx, listKey)
		if err != redis.TxFailedErr {
			return err
		}
	}
	
	return nil
}

func Get() (string, error) {
	res, err := client.LRange(listKey, 0, 0).Result()
	if err != nil {
		return "", err
	}
	return res[0], nil
}
