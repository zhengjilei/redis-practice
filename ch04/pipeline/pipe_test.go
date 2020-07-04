package pipeline

import (
	"fmt"
	"github.com/zhengjilei/redis-practice/common/redisclient"
	"testing"
	"time"
)

func BenchmarkNormal(b *testing.B) {
	client := redisclient.Client
	defer client.Close()
	for i := 0; i < b.N; i++ {
		client.Set(fmt.Sprintf("abcde:1234:%d", i), 999, 10*time.Second)
		client.Set(fmt.Sprintf("abcde:32:%d:%d", i, i), 1000, 100*time.Second)
		client.HSet(fmt.Sprintf("user-profile:%d:1234", i), "username", "king foo")
		client.HSet(fmt.Sprintf("user-session:%d:1234", i), "username", "king foo")
	}
}

func BenchmarkTestPipeline(b *testing.B) {
	client := redisclient.Client
	defer client.Close()

	pipe := client.Pipeline()
	for i := 0; i < b.N; i++ {
		pipe.Set(fmt.Sprintf("abcde:1234:%d", i), 999, 10*time.Second)
		pipe.Set(fmt.Sprintf("abcde:32:%d:%d", i, i), 1000, 100*time.Second)
		pipe.HSet(fmt.Sprintf("user-profile:%d:1234", i), "username", "king foo")
		pipe.HSet(fmt.Sprintf("user-session:%d:1234", i), "username", "king foo")
		pipe.Exec()
	}
}
