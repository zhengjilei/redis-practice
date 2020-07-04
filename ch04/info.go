package main

import (
	"fmt"
	"log"
	"github.com/zhengjilei/redis-practice/common/redisclient"
	"strings"
)

func main() {

	res, err := redisclient.Client.Info("MEMORY").Result()
	if err != nil {
		log.Fatalln(err)
	}
	str := strings.Split(res, "\n")
	paramMap := map[string]string{}
	for _, s := range str {
		ss := strings.Split(s, ":")
		if len(ss) != 2 {
			continue
		}
		paramMap[ss[0]] = ss[1]
	}
	fmt.Println(paramMap["used_memory_human"])
}
