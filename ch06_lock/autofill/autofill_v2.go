package autofill

import (
	"github.com/go-redis/redis/v7"
	"github.com/zhengjilei/redis-practice/common/redisclient"
	"os/exec"
)

const (
	retryCount = 10
	maxNum     = 20
)

func initGroup(group string, emails []string) {
	pipe := redisclient.Client.Pipeline()
	data := []*redis.Z{}
	for _, e := range emails {
		data = append(data, &redis.Z{
			Member: e,
			Score:  0,
		})
	}
	pipe.ZAdd("email:"+group, data...)
	pipe.Exec()
}

// 前提: 邮箱没有特殊字符、数字，仅由小写子母组成

func autoFillV2(group, prefix string) []string {

	// 1. get previous and after helper
	previous, after := getPreAndAfter(prefix)

	identifier := getUUID()
	previous += identifier
	after += identifier

	// 2. insert two helpers into email zset

	key := "email:" + group
	pipe := redisclient.Client.Pipeline()
	pipe.ZAdd(key, &redis.Z{
		Member: previous,
		Score:  0,
	})

	pipe.ZAdd(key, &redis.Z{
		Member: after,
		Score:  0,
	})

	pipe.Exec()

	res := []string{}
	// 3. a-get rank of previous and after, b-get matched email, c-remove two helpers
	for i := 0; i < retryCount; i++ {
		err := redisclient.Client.Watch(func(tx *redis.Tx) error {
			first := tx.ZRank(key, previous).Val()
			second := tx.ZRank(key, after).Val()

			// skip the two end
			first++
			second--

			// 只取前20个
			if second > first+maxNum-1 {
				second = first + maxNum - 1
			}
			var err error
			res, err = tx.ZRange(key, first, second).Result()
			if err != nil {
				return err
			}
			pipe := tx.Pipeline()
			pipe.ZRem(key, previous)
			pipe.ZRem(key, after)
			pipe.Exec()
			return nil
		}, key)
		if err != nil {
			if err != redis.TxFailedErr {
				return []string{}
			}
			// continue retry
		} else {
			// get result
			break
		}
	}
	return res
}

// abc 的前一个英语单词应该是 abbzz...zzz, +1 -> abb{, 故 abb{ 一定在 abc 前，且在所有ab开头的单词之后
// abc 开头的最后一个单词格式时 abczzz...zzz, +1 -> abc{, 故 abc{ 一定在所有以 abc 开头的单词之后

// abc 前一个插入元素应该是 abb{, aba 前一个插入元素应该是 ab`(可以加上{, 即 ab`{)
// abc 后一个插入元素应该是 abc{, abz 后一个插入元素应该是 abz{
func getPreAndAfter(s string) (string, string) {

	c := s[len(s)-1] // 取得最后一个元素

	// get previous: 前n-1位,倒数最后一位变成子母的前序，并且加上{
	pre := s[:len(s)-1] + string(c-1) + "{"

	// get after: 直接在后面拼接一个 {
	aft := s + "{"
	return pre, aft
}
func getUUID() string {
	out, _ := exec.Command("uuidgen").Output()
	return string(out)
}
