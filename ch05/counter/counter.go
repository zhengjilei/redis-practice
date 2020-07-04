package counter

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/zhengjilei/redis-practice/common/redisclient"
	"sort"
	"strconv"
	"strings"
	"time"
)

// hash  counter:5:hits   key 是5s时间片开始的时间戳，val 是这 5 s 的点击量; 例如  <15,12> 表示 15s~20s 点击量是12
// zset  known   members 是计数器类型(5:hits 60:hits)  score 均为 0, 故该 zset 是按照member 排序，用 zset 是能按照稳定的顺序来扫描所有的 member

const (
	SAMPLE_COUNT = 3
)

var (
	precision = []int64{1, 5, 60, 300, 3600}
)

// 给每个精度统计集 加上 count
func updateCounter(count int) {
	sec := time.Now().Unix()

	pipe := redisclient.Client.Pipeline()

	for _, prec := range precision {
		ts := (sec / prec) * prec
		pipe.HIncrBy(fmt.Sprintf("counter:%d:hits", prec), strconv.FormatInt(ts, 10), int64(count))
		pipe.ZAdd("known", &redis.Z{
			Score:  0,
			Member: fmt.Sprintf("%d:hits", prec),
		})
	}
	pipe.Exec()
}

// timestamp, count
func getCounter(prec int) ([]int64, []int64, error) {
	counter, err := redisclient.Client.HGetAll(fmt.Sprintf("counter:%d:hits", prec)).Result()
	if err != nil {
		return nil, nil, err
	}
	keys := []int64{}
	for k := range counter {
		ts, _ := strconv.Atoi(k)
		keys = append(keys, int64(ts))
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	vals := []int64{}
	for _, k := range keys {
		num, _ := strconv.Atoi(counter[strconv.Itoa(int(k))])
		vals = append(vals, int64(num))
	}
	return keys, vals, nil
}

// 周期性地对计数器进行清理
// 清理时有几个问题需要注意：
// 1. 任何时候都有可能有新的计数器添加进来 => 清理时注意考虑可能需要用到 watch 的地方
// 2. 同一时间可能需要清理多种 精度的计数
// 3. 按照不同频率对不同精度的计数器进行清理，比如：每分钟清理一次计数器，对于每天只更新一次的计数器来说，就不需要
// 4. 如果计数器不包含数据，则不要清理

// 实现思路:
// a. 程序每隔1分钟执行一次清理
// b. 清理时根据计数器的精度判断此次是否需要清理，不需要则跳过
// c. 需要清理的计数器: 1. 清理member中该计数器 2. 清理对应的计数器

func cleaner() {
	cycle := 0
	for {

		start := time.Now().Unix()
		// clean

		index := int64(0)
		for ; index < redisclient.Client.ZCard("known").Val(); index++ {
			counter, err := redisclient.Client.ZRange("known", index, index).Result()
			if err != nil || len(counter) != 1 {
				continue
			}
			hits := counter[0] // 60:hits
			hitss := strings.Split(hits, ":")
			prec, _ := strconv.Atoi(hitss[0])
			bprec := prec / 60
			if !(bprec == 0 || cycle%bprec == 0) {
				// bprec ==0 保证精度小于 60 的每次循环都能执行清理
				// cycle % bprec ==0 保证首次循环每个精度的计数器都执行清理，且每隔一定时间各个精度都能准时清理
				continue
			}

			// 得到待清理的 Hash 所有的 key
			hkey := fmt.Sprintf("counter:%s", hits)
			keys := redisclient.Client.HKeys(hkey).Val()
			sort.Strings(keys)

			// 计算要删除的键总数, 只保留当前时刻最邻近的 SAMPLE_COUNT 个时间间隔
			cutoffTs := start - int64(SAMPLE_COUNT*prec)
			removeCount := 0 //
			for _, v := range keys {
				ts, _ := strconv.Atoi(v)
				if int64(ts) < cutoffTs {
					removeCount++
				}
			}

			// 删除
			if removeCount > 0 {
				redisclient.Client.HDel(hkey, keys[:removeCount]...)
				if removeCount == len(keys) { // 该更新品率的计数器已经被清空
					redisclient.Client.Watch(func(tx *redis.Tx) error {
						if tx.HLen(hkey).Val() != 0 {
							tx.Unwatch(hkey)
							return err
						}
						tx.ZRem("known", hits)
						index--
						return nil
					}, hkey)
				}
			}
		}

		end := time.Now().Unix()
		// 休眠若干秒，保证清理动作是每一分钟执行一次
		cycle++
		sleepSeconds := int64(60 - (end - start))
		if sleepSeconds <= 0 {
			continue
		}
		time.Sleep(time.Duration(sleepSeconds) * time.Second)
	}
}
