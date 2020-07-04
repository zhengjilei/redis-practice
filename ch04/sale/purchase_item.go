package main

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/zhengjilei/redis-practice/ch06_lock/lockutil"
	"strconv"
	"time"
)

const (
	hashFundKey = "funds"
)

func purchaseItem(buyerID, sellerID int64, item string) error {
	// 1. 判断 market 是否仍然有该商品，没有则直接退出
	// 2. 获取买家的余额
	// 3. 获取要买商品的价值，比较余额是否大于商品价值
	// 4. 购买：(1) 扣除买家的余额 (2)增加卖家的余额 (3) 商品加到买家的 inventory (4) 商品从 market 移除

	// 当对一个 key setting 之前，对其中的数据有严格的依赖（余额要够，商品要存在），需要watch
	
	//需要 watch 的：依赖读，再写的，需要进行 watch
	// market(减商品)：保证商品还在, buyer(减余额): 保证有足够的余额
	// 不需要 watch 的: buyerInventory(加商品), seller(加余额），ß只是一个incr、add 的操作，没有数据严格的依赖

	itemKey := getMarketItemKey(item, sellerID)
	buyerKey := getUserKey(buyerID)
	sellerKey := getUserKey(sellerID)
	buyerInventoryKey := getInventoryKey(buyerID)

	end := time.Now().Add(5 * time.Second)

	for retries := maxRetryCount; time.Now().Before(end) && retries > 0; retries-- {
		ftx := func(tx *redis.Tx) error {
			sellPrice, err := tx.ZScore(marketKey, itemKey).Result()
			if err != nil {
				//tx.Unwatch(marketKey, buyerKey, )
				if err == redis.Nil {
					return fmt.Errorf("item:%s was already not in market", itemKey)
				}
				return err
			}
			// ensure the balance is enough to afford the item
			if balanceStr, err := tx.HGet(buyerKey, hashFundKey).Result(); err == nil {
				if balance, err := strconv.ParseFloat(balanceStr, 64); err == nil {
					if sellPrice > balance {
						return errors.New("no enough money")
					}

					// runs only if the watched keys remain unchanged
					_, err = tx.Pipelined(func(pipe redis.Pipeliner) error {
						pipe.HIncrByFloat(sellerKey, hashFundKey, sellPrice)
						pipe.HIncrByFloat(buyerKey, hashFundKey, -sellPrice)
						pipe.ZRem(marketKey, itemKey)
						pipe.SAdd(buyerInventoryKey, item)
						return nil
					})
				}
				return err
			}

			return err
		}
		err := client.Watch(ftx, marketKey, buyerKey)
		if err != redis.TxFailedErr {
			return err
		}
	}
	return nil
}

func purchaseItemWithLock(buyerID, sellerID int64, item string) error {
	itemKey := getMarketItemKey(item, sellerID)
	buyerKey := getUserKey(buyerID)
	sellerKey := getUserKey(sellerID)
	buyerInventoryKey := getInventoryKey(buyerID)

	// 粗粒度锁: 锁住了 market，任何想要操作 market 的数据，都必须先获得 market 的锁
	// 细粒度锁: 只锁住 market 中要购买的 item    lockName = marketKey+":"+itemKey
	lockID, err := lockutil.AcquireLock(client, marketKey, 10)
	if err != nil {
		return err
	}
	defer lockutil.ReleaseLock(client, marketKey, lockID)

	result, err := client.Pipelined(func(pipe redis.Pipeliner) error {
		client.ZScore(marketKey, itemKey)
		client.HGet(buyerKey, hashFundKey)
		return nil
	})

	var sellPrice float64
	if c, ok := result[0].(*redis.FloatCmd); ok {
		sellPrice, err = c.Result()
		if err == redis.Nil {
			return fmt.Errorf("item:%s was already not in market", itemKey)
		}
		if err != nil {
			return err
		}
	}
	var balance float64
	if c, ok := result[1].(*redis.FloatCmd); ok {
		balance, err = c.Result()
		if err != nil {
			return err
		}
	}
	if balance < sellPrice {
		return errors.New("balance is not enough")
	}

	_, err = client.TxPipelined(func(pipe redis.Pipeliner) error {
		pipe.HIncrByFloat(sellerKey, hashFundKey, sellPrice)
		pipe.HIncrByFloat(buyerKey, hashFundKey, -sellPrice)
		pipe.ZRem(marketKey, itemKey)
		pipe.SAdd(buyerInventoryKey, item)
		return nil
	})
	return err
}
