package main

import "github.com/zhengjilei/redis-practice/common/redisclient"

var client = redisclient.Client

func main() {
	logRecent()
}
