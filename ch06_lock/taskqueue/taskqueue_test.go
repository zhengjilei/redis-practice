package taskqueue

import (
	"github.com/zhengjilei/redis-practice/common/redisclient"
	"log"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	cancel := make(chan struct{})
	var runPET = func(id string) {
		log.Println("pet ", id, " start...")
		time.Sleep(10 * time.Second)
		log.Println("pet ", id, " end...")
	}
	wq := NewWorkerQueue(redisclient.Client, []string{"pet:high", "pet:medium", "pet:low"}, cancel, runPET, 3, "pet")
	wq.Start()
	<-time.After(5 * time.Second)

	wq.AddTask("pet:medium", 100, 1)
	wq.AddTask("pet:medium", 101, 3)

	time.Sleep(5 * time.Second)
	cancel <- struct{}{}
	<-time.After(30 * time.Second)
}
