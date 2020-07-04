package ch02

import (
	"github.com/go-redis/redis/v7"
	"math"
	"github.com/zhengjilei/redis-practice/common/redisclient"
	"strconv"
	"time"
)

const (
	MaxLimitSessions = 10000000
)

// hash: key=login,subkey={token},subvalue={userID}  记录token 和 userID 的映射关系
// zset: key=recent, member={token},score={timestamp}   记录最近访问的所有会话 token，维持会话中始终保持1000万个
// zset: key=viewed:{token},member={item},score={timestamp}  记录某个会话的访问商品历史，只保留最近访问的25个
// 购物车 hash: key=cart:{token} subkey={item} subvalue={count}

// 改进：viewed:{token} 集合中保存了该用户 每个商品浏览的时间，只在删除时用到了，保存最近25个商品，可以用列表代替，减少因保存时间戳占用的空间
// l_viewed:{token}
// 1. 新访问一个商品,rpush 到list 的右边
// 2. ltrim (-25,-1)  只保留最近 25个元素

func checkToken(token string) (int, error) {
	userID, err := redisclient.Client.HGet("login", token).Result()
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(userID)
}

// 用 list 代替 zset
func UpdateTokenV2(token string, userID int, item string) (bool, error) {
	// 记录token 和 user 之间的映射关系
	ok, err := redisclient.Client.HSet("login", token, strconv.Itoa(userID)).Result()
	if err != nil {
		return false, err
	}
	if ok == 0 {
		return false, nil
	}

	redisclient.Client.RPush("l_recent", token)

	if "" != item {
		redisclient.Client.RPush("l_viewed:"+token, item)
		redisclient.Client.LTrim("l_viewed:"+token, -25, -1)
	}
	return true, nil
}

func UpdateToken(token string, userID int, item string) (bool, error) {
	timestamp := time.Now().UTC().Unix()
	// 记录token 和 user 之间的映射关系
	ok, err := redisclient.Client.HSet("login", token, strconv.Itoa(userID)).Result()
	if err != nil {
		return false, err
	}
	if ok == 0 {
		return false, nil
	}
	// 记录 token 最后一次登录的会话时间戳
	redisclient.Client.ZAdd("recent", &redis.Z{
		Score:  float64(timestamp),
		Member: token,
	})

	if "" != item {
		// 添加到该用户浏览商品历史记录上
		redisclient.Client.ZAdd("viewed:"+token, &redis.Z{
			Score:  float64(timestamp),
			Member: item,
		})
		redisclient.Client.ZRemRangeByRank("viewed:"+token, 0, -26) // 只保留最近浏览的25项商品（timestamp 最大的25个）
	}
	return true, nil
}

func AddToCart(token string, itemID, count int64) {
	if count <= 0 {
		redisclient.Client.HDel("cart:"+token, strconv.FormatInt(itemID, 10))
	} else {
		redisclient.Client.HSet("cart:"+token, strconv.FormatInt(itemID, 10), count)
	}
}

// 定期清理会话，保持会话数维持在 1000,000 个
func clearSessions() error {
	for {
		size, err := redisclient.Client.ZCard("recent").Result()
		if err != nil {
			panic(err)
		}
		if size <= MaxLimitSessions {
			time.Sleep(time.Second * 1)
			continue
		}
		// 每次清理最旧的 100 个
		// size = 101 max = 100 delete [0,101-100-1]
		endIndex := int64(math.Min(float64(size-MaxLimitSessions-1), 99))
		tokens, err := redisclient.Client.ZRange("recent", 0, endIndex).Result()
		if err != nil {
			return err
		}

		redisclient.Client.HDel("login", tokens...)

		var tokensInterfaces []interface{}
		var viewedTokenKeys []string
		var cartTokenKeys []string
		for _, v := range tokens {
			tokensInterfaces = append(tokensInterfaces, v)
			viewedTokenKeys = append(viewedTokenKeys, "viewed:"+v)
			cartTokenKeys = append(cartTokenKeys, "cart:"+v)
		}
		redisclient.Client.ZRem("recent", tokensInterfaces...)
		redisclient.Client.Del(viewedTokenKeys...)

		// clear cart
		redisclient.Client.Del(cartTokenKeys...)
	}
}

func clearSessionsV2() error {
	for {
		time.Sleep(time.Second * 1)
		size, err := redisclient.Client.LLen("l_recent").Result()
		if err != nil {
			continue
		}
		if size <= MaxLimitSessions {
			continue
		}
		// 清理最旧的 100 个 token
		tokens, err := redisclient.Client.LRange("l_recent", 0, 99).Result()
		if err != nil {
			continue
		}
		// 保留倒数第一个-> 倒数第 size-100个
		redisclient.Client.LTrim("l_recent", 100-size, -1)

		redisclient.Client.HDel("login", tokens...)

		var viewedTokenKeys []string
		var cartTokenKeys []string
		for _, v := range tokens {
			viewedTokenKeys = append(viewedTokenKeys, "l_viewed:"+v)
			cartTokenKeys = append(cartTokenKeys, "cart:"+v)
		}
		redisclient.Client.Del(viewedTokenKeys...)
		redisclient.Client.Del(cartTokenKeys...)

	}
}
