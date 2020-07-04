package tt

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/zhengjilei/redis-practice/common/redisclient"
	"github.com/zhengjilei/redis-practice/common/util"
	"testing"
	"time"
)

func TestBlpop(t *testing.T) {
	client := redisclient.Client
	defer client.Close()

	str, err := client.BLPop(5*time.Second, "a", "b").Result()
	fmt.Println(err)
	fmt.Println(str)
}
func TestTTL(t *testing.T) {
	client := redisclient.Client
	defer client.Close()
	testCases := []struct {
		desc     string
		testFunc func()
	}{
		{
			desc: "normal",
			testFunc: func() {
				ttl, err := client.TTL("c").Result() // -1,nil
				if err != nil {
					panic(err)
				}
				fmt.Println(int(ttl))

				ttl2, err := client.TTL("cd").Result() // -2,nil
				if err != nil {
					panic(err)
				}
				fmt.Println(int(ttl2)) // -2

			},
		},
	}
	for _, testCase := range testCases {
		testCase.testFunc()
	}
}

func TestPipeline(t *testing.T) {

	client := redisclient.Client
	defer client.Close()

	result, err := client.Pipelined(func(pipe redis.Pipeliner) error {
		pipe.ZAdd("mzset", &redis.Z{
			Score:  12,
			Member: "abc",
		})
		pipe.LPush("mlist", "cde")
		pipe.Set("cc", 32, -1)
		pipe.ZRangeWithScores("mzset", 0, -1)
		return nil
	})
	fmt.Println("err:", err)
	util.PrintResult(result)

	res, ok := result[0].(*redis.IntCmd)
	fmt.Println("1: ok= ", ok)
	fmt.Println(res.Val())

	res, ok = result[1].(*redis.IntCmd)
	fmt.Println("2: ok= ", ok)
	fmt.Println(res.Val())

	res2, ok := result[2].(*redis.StatusCmd)
	fmt.Println("3: ok= ", ok)
	fmt.Println(res2.Val())

	res3, ok := result[3].(*redis.ZSliceCmd)
	fmt.Println("4: ok= ", ok)
	for _, v := range res3.Val() {
		fmt.Println(v.Member, ":", v.Score)
	}

}

func TestRemZSetKey(t *testing.T) {
	mzetKey := "mzset"

	redisclient.Client.ZAdd(mzetKey, &redis.Z{
		Score:  21,
		Member: "a",
	})

	res, err := redisclient.Client.ZRem(mzetKey, "a").Result()
	fmt.Println(res)
	fmt.Println(err)
	fmt.Println("-----------")
	res, err = redisclient.Client.ZRem(mzetKey, "a").Result()
	fmt.Println(res)
	fmt.Println(err)
	fmt.Println("-----------")
	res, err = redisclient.Client.ZRem("non-exist-key", "a").Result()
	fmt.Println(res)
	fmt.Println(err)
}
