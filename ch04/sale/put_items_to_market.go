package main

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"time"
)

const (
	maxRetryCount = 1000

	marketKey = "market"
)

// user hash , i.e.  users:21----{name=ethan,funds=21}
func getUserKey(userID int64) string {
	return fmt.Sprintf("users:%d", userID)
}

// market hash item, i.e. itemA.12, belong to  markets----{"itemsA.12"=35,"items.B"=32}
func getMarketItemKey(item string, sellerID int64) string {
	return fmt.Sprintf("%s.%d", item, sellerID)
}

// Inventory set i.e.  inventory:27----[itemA,itemB]
func getInventoryKey(sellerID int64) string {
	return fmt.Sprintf("inventory:%d", sellerID)
}

func putIntoMarket( item string, sellerID int64, price float64) error {
	inventoryKey := getInventoryKey(sellerID)
	itemKey := getMarketItemKey(item, sellerID)
	endTime := time.Now().Add(time.Second * 3)
	txf := func(tx *redis.Tx) error {
		// 立即返回结果
		flag, err := tx.SIsMember(inventoryKey, item).Result()
		if err != nil || !flag {
			if err != nil {
				fmt.Println("ismember err: ", err)
			}
			// tx Close的时候会自动执行 unwatch
			//tx.Unwatch(inventoryKey)
			return fmt.Errorf("item:%s is already not in %s", itemKey, inventoryKey)
		}

		// transaction operation，一般是 setting 操作

		// runs only if the watched keys remain unchanged
		result, err := tx.Pipelined(func(pipe redis.Pipeliner) error {
			// pipe handles the error case
			pipe.ZAdd(marketKey, &redis.Z{
				Score:  price,
				Member: itemKey,
			})
			pipe.SRem(inventoryKey, item)
			return nil
		})
		for _, v := range result {
			fmt.Println(v.Name(), v.Args(), v.Err())
		}
		return err
	}
	for retries := maxRetryCount; time.Now().Before(endTime) && retries > 0; retries-- {
		err := client.Watch(txf, inventoryKey)
		if err != redis.TxFailedErr {
			// nil or item is already not in inventor set
			fmt.Println(err)
			return err
		}

		// txFailed Err, maybe caused by watch key: inventor was changed by other client 
		// will retry
	}
	return errors.New("failed to add to market, reach to maximum retry count or timeoutß")
}

func hello() {
	fmt.Println("hello")
}
