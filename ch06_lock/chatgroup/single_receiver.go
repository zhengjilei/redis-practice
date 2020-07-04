package chatgroup

import "github.com/zhengjilei/redis-practice/common/redisclient"

const (
	mailPrefix = "mailbox"
)

func sendMsg(receiver, msg string) (bool, error) {
	count, err := redisclient.Client.RPush(mailPrefix+":"+receiver, msg).Result()
	if err != nil {
		return false, err
	}
	return count == 1, nil
}
