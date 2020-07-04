package chatgroup

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v7"
	"log"
	"github.com/zhengjilei/redis-practice/ch06_lock/lockutil"
	"github.com/zhengjilei/redis-practice/common/redisclient"
	"time"
)

const (
	chatIncrKey = "ids:chat"
	msgIncrKey  = "ids:msg"
)

func getChatZSetKey(chatID int64) string {
	return fmt.Sprintf("chat:%d", chatID)
}

func getUserZSetKey(username string) string {
	return fmt.Sprintf("seen:%s", username)
}

func getMsgZSetKey(chatID int64) string {
	return fmt.Sprintf("msgs:%d", chatID)
}

type Message struct {
	ID        int64  `json:"id"`
	Timestamp string `json:"timestamp"`
	Sender    string `json:"sender"`
	Content   string `json:"content"`
}

func sendMessages(sender, msg string, chatID int64) error {
	identifier, err := lockutil.AcquireLockV3(redisclient.Client, getChatZSetKey(chatID), 3, 10)
	if err != nil {
		return err
	}
	msgID := redisclient.Client.Incr(identifier).Val()

	msgObj := &Message{
		ID:        msgID,
		Timestamp: time.Now().String(),
		Sender:    sender,
		Content:   msg,
	}
	data, err := json.Marshal(msgObj)
	if err != nil {
		log.Println(err)
		return err
	}
	err = redisclient.Client.ZAdd(getMsgZSetKey(chatID), &redis.Z{
		Score:  float64(msgID),
		Member: data,
	}).Err()

	if err != nil {
		log.Println(err)
		return err
	}
	// TODO 更新已读 MSG ID? 
	return nil
}
func createChat(organizer string, users []string, initMsg string) error {
	// 获得最新的 chat id
	chatID := redisclient.Client.Incr(chatIncrKey).Val()

	// 1. 将所有 user 加到 chat:{id} 这个 zset 集合内
	// 2. 将每个 user 创建一个新的 seen:{username} zset 集合，记录对应 chat 的阅读进度
	_, err := redisclient.Client.Pipelined(func(pipe redis.Pipeliner) error {
		for _, user := range users {
			pipe.ZAdd(getChatZSetKey(chatID), &redis.Z{
				Score:  0,
				Member: user,
			})
			pipe.ZAdd(getUserZSetKey(user), &redis.Z{
				Score:  0,
				Member: getChatZSetKey(chatID),
			})
		}
		return nil
	})

	if err != nil {
		return err
	}
	// 发送一条初始信息
	return sendMessages(organizer, initMsg, chatID)
}

func fetchPendingMsg(user string) ([]*Message, error) {
	client := redisclient.Client
	pipe := client.Pipeline()
	// 1. 获取当前用户的zset, 得到用户在各个 chat 的已读位置
	result := pipe.ZRangeWithScores(getUserZSetKey(user), 0, -1).Val()

	var messages []*Message
	// 2. 遍历该用户在每个chat关联的msg，取到 [已读编号+1,+inf] 的所有数据
	for _, z := range result {
		chatID := z.Member
		maxReadSeq := z.Score

		result, err := pipe.ZRangeByScoreWithScores(getMsgZSetKey(chatID.(int64)), &redis.ZRangeBy{
			Min: fmt.Sprintf("%d", int(maxReadSeq)+1),
			Max: "+inf",
		}).Result()
		if err != nil {
			log.Println(err)
			return nil, err
		}
		for _, z := range result {
			msg := Message{}
			if err := json.Unmarshal(z.Member.([]byte), &msg); err != nil {
				log.Println(err)
				continue
			}
			messages = append(messages, &msg)
		}
		//TODO .....

	}
	// 3. 更新该用户zset的各个chat 的已读位置，更新该用户所在的 chat 的zset 已读位置

	// 4. 遍历该用户所在的 chat，删除每个 chat 中被所有人已经收到的信息（得到每个 chat zset 的最小分值，删除对应的 msg zset 里最小分值及其之前的数据）
	_, _ = pipe.Exec()

}
