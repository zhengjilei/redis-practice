package main

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"log"
	"github.com/zhengjilei/redis-practice/common/redisclient"
	"testing"
	"time"
)

func TestLua(t *testing.T) {
	rdb := redisclient.Client

	IncrByXX := redis.NewScript(`
		if redis.call("GET", KEYS[1]) ~= false then
			return redis.call("INCRBY", KEYS[1], ARGV[1])
		end
		return false
	`)

	n, err := IncrByXX.Run(rdb, []string{"xx_counter"}, 2).Result()
	fmt.Println(n, err) // <nil> redis: nil

	_, ok := n.(*redis.BoolCmd)
	fmt.Println(ok) // false

	err = rdb.Set("xx_counter", "40", 0).Err()
	if err != nil {
		panic(err)
	}

	n, err = IncrByXX.Run(rdb, []string{"xx_counter"}, 2).Result()
	fmt.Println(n, err) // 42 <nil>
}

func TestLua2(t *testing.T) {
	script := ` 
				local i = tonumber(ARGV[1])
				local res
				while(i>0) do
					res = redis.call('lpush',KEYS[1],math.random())
					i = i - 1
				end
				return res
				`
	sc := redis.NewScript(script)
	num, err := sc.Run(redisclient.Client, []string{"mlist2"}, 10).Int()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(num)
}

func Test3(t *testing.T) {
	str, err := redisclient.Client.Get("dsdsadsaa").Result()
	fmt.Println(err)              // redis: nil
	fmt.Println(err == redis.Nil) // true
	fmt.Println(str)
}

// type(geohashcodekey) == table
const (
	addPOIBlockScript = `
							if redis.call('ZADD',KEYS[1],ARGV[1],ARGV[2]) == 1 then
								if redis.call('ZCARD',KEYS[1]) > tonumber(ARGV[3]) then
									local tab = redis.call('ZRANGE',KEYS[1],0,0,'withscores')
									if redis.call('DEL', tab[1] ) == 1 then
										redis.call('ZREMRANGEBYRANK',KEYS[1],0,0)
										return 2
									end
								end
							end
							return 1
						`

	str = `							if redis.call('ZCARD',KEYS[1]) > tonumber(ARGV[3]) then
									local tab = redis.call('ZRANGE',KEYS[1],0,0,'withscores')
									if redis.call('DEL', tab[1] ) == 1 then
										redis.call('ZREMRANGEBYRANK',KEYS[1],0,0)
									end
								end`
)

func Test4(t *testing.T) {

	sc := redis.NewScript(addPOIBlockScript)
	args := []interface{}{time.Now().UTC().UnixNano(), "{it}:DSA3FGD:2020-05-03", 2}
	keys := []string{"poi_block:it"}

	res, err := sc.Run(redisclient.Client, keys, args...).Result()
	_ = res
	_ = err
}
