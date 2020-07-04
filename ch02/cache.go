package ch02

import (
	"fmt"
	"log"
	"github.com/zhengjilei/redis-practice/common/redisclient"
	"net/http"
	"strconv"
	"strings"
)

func cacheRequest(req *http.Request, callback func(req *http.Request) *http.Response) *http.Response {
	if !canCache(req) {
		return callback(req)
	}
	pageKey := "cache:" + hashRequest(req)

	pageVal, err := redisclient.Client.Get(pageKey).Result()
	if err != nil {
		log.Println(err)
		return nil
	}

	if "" == pageVal {
		resp := callback(req)
		_, err := redisclient.Client.Set(pageKey, getRedisValStr(resp), -1).Result()
		if err != nil {
			log.Println(err)
		}
		return resp
	}

	return unmarshalRedisValStr(pageVal)
}

func canCache(req *http.Request) bool {
	return true
}
func hashRequest(req *http.Request) string {
	return req.Host + req.Method
}

func getRedisValStr(resp *http.Response) string {
	return fmt.Sprintf("%s:%d", resp.Status, resp.StatusCode)
}
func unmarshalRedisValStr(resp string) *http.Response {
	resps := strings.Split(resp, ":")
	statusCode, _ := strconv.Atoi(resps[1])
	return &http.Response{
		Status:     resps[0],
		StatusCode: statusCode,
	}
}
