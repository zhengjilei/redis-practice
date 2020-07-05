package taskqueue

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/zhengjilei/redis-practice/ch06_lock/lockutil"
	"log"
	"strings"
	"sync"
	"time"
)

var once sync.Once

type WorkerQueue struct {
	queues       []string
	delayedQueue string
	client       *redis.Client
	fn           func(string)
	cancel       chan struct{}
	interval     int
}

const (
	delayedQueueKey = "delayed"
)

// queues: 按照优先级从高到低的顺序,一般分为: high-priority medium-priority,low-priority
func NewWorkerQueue(client *redis.Client, queues []string, cancel chan struct{}, fn func(string), interval int, taskName string) *WorkerQueue {
	return &WorkerQueue{
		queues:       queues,
		delayedQueue: delayedQueueKey + ":" + taskName,
		fn:           fn,
		cancel:       cancel,
		interval:     interval, // 阻塞拉取队列中的元素，每隔多少秒超时重试
		client:       client,
	}
}
func (wq *WorkerQueue) Cancel() {
	wq.cancel <- struct{}{}
}

// zset: delayed:taskName      key={queueName}.{taskID}  value={timestamp}
// list: queueName      taskID
// string(lock):  lock:queueName.taskID    
func (wq *WorkerQueue) Start() {
	once.Do(func() {
		// 取任务队列中的值
		go func() {
			for ; ; {
				select {
				case <-wq.cancel:
					log.Println("goroutine[1]: canceled...")
					wq.Close()
					return
				default:
				}

				taskID, err := wq.client.BLPop(time.Duration(wq.interval)*time.Second, wq.queues...).Result()
				if err != nil {
					if err == redis.Nil {
						// timeout
						log.Println("goroutine[1]: timeout")
						continue
					}
					log.Println("goroutine[1]: error when blpop", err)
					continue
				}
				// taskID = [listName,123]
				log.Println("goroutine[1]: start ", taskID)
				wq.fn(taskID[1])
			}
		}()

		// 将定时调度的任务，转移到指定优先级的任务队列中 
		go func() {
			for ; ; {
				time.Sleep(1 * time.Second)
				nowInMilli := time.Now().UnixNano() / int64(time.Millisecond)
				res := wq.client.ZRangeWithScores(wq.delayedQueue, 0, 0).Val()
				if len(res) != 1 || res[0].Score >= float64(nowInMilli) {
					//log.Printf("goroutine[2]: now:%d res:%v\n", nowInMilli, res)
					continue
				}
				queueAndTaskID := res[0].Member.(string)
				str := strings.Split(queueAndTaskID, ".")
				if len(str) != 2 {
					// log error: invalid format
					log.Println("goroutine[2]: invalid format", queueAndTaskID)
					wq.client.ZRem(wq.delayedQueue, queueAndTaskID)
					continue
				}
				queue := str[0]
				taskID := str[1]

				log.Println("goroutine[2]: try to push ", queueAndTaskID)
				// 只是对当前任务进行加锁，防止多个客户端同时对该任务操作
				identifier, err := lockutil.AcquireLockV3(wq.client, queueAndTaskID, 5, 10)
				if err != nil {
					continue
				}

				log.Println("goroutine[2]: get lock", queueAndTaskID)
				// TODO: should use lua script to keep atomic
				delCount, _ := wq.client.ZRem(wq.delayedQueue, queueAndTaskID).Result()
				if delCount == 1 { // 说明当前获得锁的客户端该 queueAndTaskID 删除了
					wq.client.RPush(queue, taskID) // 压到 scheduled task 指定的队列
				}
				_ = lockutil.ReleaseLock(wq.client, queueAndTaskID, identifier)
			}
		}()
	})

}

func (wq *WorkerQueue) Close() {
	close(wq.cancel)
	wq.fn = nil
	wq.queues = nil
}

// 定时任务
// queue：指定scheduled task 进入哪个 queue 运行，通过 queue 可以设定该 task 的优先级
func (wq *WorkerQueue) AddTask(queue string, taskID, delayInSeconds int64) (bool, error) {
	if !wq.checkQueue(queue) {
		return false, errors.New("non-supported queue")
	}
	if delayInSeconds > 0 {
		added, err := wq.client.ZAdd(wq.delayedQueue, &redis.Z{
			Score:  float64(time.Now().UnixNano() / int64(time.Millisecond)),
			Member: fmt.Sprintf("%s.%d", queue, taskID),
		}).Result()
		if err != nil {
			log.Println("add task ", err)
			return false, err
		}
		return added == 1, nil
	}
	// add task queue instantly
	i, err := wq.client.RPush(queue, taskID).Result()
	if err != nil {
		return false, err
	}
	return i == 1, nil
}

func (wq *WorkerQueue) checkQueue(queue string) bool {
	exist := false
	for _, v := range wq.queues {
		if v == queue {
			exist = true
			break
		}
	}
	return exist
}
