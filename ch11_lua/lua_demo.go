package main

import (
	"github.com/go-redis/redis/v7"
	"strings"
)

var (
	LuaCallFunc func(client *redis.Client, keys, args []string, forceEval bool) error
)

// 加载 script，返回一个函数，下一次执行直接调用函数就是直接 EvalSha()
// 可以直接通过 func (s *Script) Run(） 代替
func ScriptLoad(script string) interface{} {
	sha := ""
	call := func(client *redis.Client, keys, args []string, forceEval bool) error {
		if !forceEval {
			// 不强制使用 Eval 尽量使用 EvalSHA
			if len(sha) == 0 {
				// 只在第一次函数被调用时会执行load 脚本, 得到 sha 值
				//script load 
				sha = client.ScriptLoad(script).Val()
			}
			err := client.EvalSha(sha, keys, args).Err()
			if err != nil && strings.HasPrefix(err.Error(), "NOSCRIPT") {
				// do nothing, to execute Eval
			} else {
				return err
			}
		}
		// 强制使用 Eval
		return client.Eval(script, keys, args).Err()
	}
	return call
}
