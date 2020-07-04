package util

import (
	"fmt"
	"github.com/go-redis/redis/v7"
)

func PrintResult(res []redis.Cmder) {
	for _, v := range res {
		fmt.Println(v.Name(), v.Args(), v.Err())
	}
}
