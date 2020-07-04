package autofill

import (
	"github.com/zhengjilei/redis-practice/common/redisclient"
	"strings"
)

// 适用于小数据量，将所有的数据一次性取出来，在内存中进行过滤

// 只保留最近 100 个联系人
func upsertContact(user, name string) {
	pipe := redisclient.Client.Pipeline()
	key := "recent:" + user
	pipe.LRem(key, 1, name)
	pipe.LPush(key, name)
	pipe.LTrim(key, 0, 99)
	pipe.Exec()
}

// 从列表里移除一个元素的时间复杂度是 O(n), 即和列表长度成正比
func removeContact(user, name string) {
	key := "recent:" + user
	redisclient.Client.LRem(key, 1, name)
}

func autoFill(user, input string) []string {
	key := "recent:" + user
	strs, err := redisclient.Client.LRange(key, 0, -1).Result()
	if err != nil {
		return []string{}
	}
	match := []string{}
	lowerInput := strings.ToLower(input)
	for _, v := range strs {
		if strings.Contains(strings.ToLower(v), lowerInput) {
			match = append(match, v)
		}
	}
	return match
}
