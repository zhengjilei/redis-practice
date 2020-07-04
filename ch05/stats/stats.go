package stats

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/zhengjilei/redis-practice/common/redisclient"
	"os/exec"
	"strconv"
	"time"
)

// zset stats:ProfilePage:AccessTime
// min 1
// max 21
// sum 44
// sumsq
// count 12

// key stats:ProfilePage:AccessTime:start 当前小时

// value 在这里指的是此次访问页面的耗时
func updateStats(category, typ string, value float64, timeout int64) error {
	destination := fmt.Sprintf("stats:%s:%s", category, typ)
	startKey := destination + ":start"
	end := time.Now().Unix() + timeout

	for ; time.Now().Unix() < end; {
		txFun := func(tx *redis.Tx) error {
			startHour, err := tx.Get(startKey).Result()
			if err != nil {
				if err != redis.Nil {
					return err
				}
				// redis.Nil execute following steps
			}
			pipe := tx.Pipeline()

			nowHour := time.Now().Hour()

			startHourI, _ := strconv.ParseInt(startHour, 10, 64)
			if startHourI < int64(nowHour) {
				pipe.Rename(destination, destination+":last")
				pipe.Rename(startKey, startKey+":pstart")
				pipe.Set(startKey, strconv.Itoa(nowHour), -1)
			}

			// 用临时的zset 记录 min max 值，方便和已有的 min,max 做比较
			tkey1 := getUUID()
			tkey2 := getUUID()
			pipe.ZAdd(tkey1, &redis.Z{
				Member: "min",
				Score:  value,
			})
			pipe.ZAdd(tkey2, &redis.Z{
				Member: "max",
				Score:  value,
			})

			pipe.ZUnionStore(destination, &redis.ZStore{
				Keys:      []string{destination, tkey1},
				Weights:   nil,
				Aggregate: "min",
			})
			pipe.ZUnionStore(destination, &redis.ZStore{
				Keys:      []string{destination, tkey2},
				Weights:   nil,
				Aggregate: "max",
			})

			pipe.Del(tkey1, tkey2)
			pipe.ZIncrBy(destination, 1, "count")
			pipe.ZIncrBy(destination, value, "sum")
			pipe.ZIncrBy(destination, value*value, "sumsq")
			pipe.Exec()
			return nil
		}
		if err := redisclient.Client.Watch(txFun, startKey); err != redis.TxFailedErr {
			return err
		}
	}
	return nil
}
func getStats(category, typ string) (map[string]float64, error) {
	key := fmt.Sprintf("stats:%s:%s", category, typ)
	res, err := redisclient.Client.ZRangeWithScores(key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	m := map[string]float64{}
	for _, v := range res {
		m[v.Member.(string)] = v.Score
	}

	if m["count"] != 0 {
		m["avg"] = m["sum"] / m["count"]
	}
	return m, nil
}

func getUUID() string {
	out, _ := exec.Command("uuidgen").Output()
	return string(out)
}
