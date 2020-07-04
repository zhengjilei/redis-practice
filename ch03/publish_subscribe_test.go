package test

import (
	"fmt"
	"github.com/zhengjilei/redis-practice/common/redisclient"
	"testing"
	"time"
)

func TestExample(t *testing.T) {
	rdb := redisclient.Client
	pubsub := rdb.Subscribe("mychannel1")

	// Wait for confirmation that subscription is created before publishing anything.
	_, err := pubsub.Receive()
	if err != nil {
		panic(err)
	}

	// Go channel which receives messages.
	ch := pubsub.Channel()

	// Publish a message.
	err = rdb.Publish("mychannel1", "hello").Err()
	if err != nil {
		panic(err)
	}

	err = rdb.Publish("mychannel1", "world").Err()
	if err != nil {
		panic(err)
	}

	time.AfterFunc(time.Second, func() {
		// When pubsub is closed channel is closed too.
		_ = pubsub.Close()
	})

	// Consume messages.
	for msg := range ch {
		fmt.Println(msg.Channel, msg.Payload)
	}

}

func TestDemo(t *testing.T) {
	go func() { subscribe() }()
	publish(10)
}
func subscribe() {
	s := redisclient.Client.Subscribe("channel")
	defer s.Close()
	//_, err := s.Receive()
	//if err != nil {
	//	panic(err)
	//}

	ch := s.Channel()
	for msg := range ch {
		fmt.Println(msg.Channel, msg.Payload)
	}
}

func publish(n int) {
	time.Sleep(1 * time.Second)
	for i := 0; i < n; i++ {
		redisclient.Client.Publish("channel", i)
		time.Sleep(1 * time.Second)
	}
}
