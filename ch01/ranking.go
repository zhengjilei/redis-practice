package ch01

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/pkg/errors"
	"log"
	"github.com/zhengjilei/redis-practice/common/redisclient"
	"strconv"
	"time"
)

type Article struct {
	ID          int64
	Title       string
	Link        string
	Score       int64
	PublishTime time.Time
}
type User struct {
	ID   int64
	Name string
}

const (
	OneWeekInSeconds      = 7 * 24 * 60 * 60
	VoteIncrScore         = 432
	VotedPrefix           = "voted"
	AgainstVotePrefix     = "against_voted"
	TimeZSetK             = "time"
	ScoreZSetK            = "score"
	VotedCountHSetK       = "voted_count"
	AgainstVoteCountHSetK = "against:voted_count"
	ArticleCountK         = "article:count"
	ArticlePrefix         = "article"
	GroupPrefix           = "group"
)

var (
	ErrTimeOut = fmt.Errorf("article cannot be voted because over 7 days since published")
)

func timeout(articleID int64) bool {
	t := time.Now().UTC().Unix() - OneWeekInSeconds
	publishTime, err := redisclient.Client.ZScore(TimeZSetK, strconv.FormatInt(articleID, 10)).Result()
	if err == redis.Nil {
		fmt.Println("zset key -- time not exist")
	} else if err != nil {
		panic(err)
	}
	if t > int64(publishTime) {
		return true
	}
	return false
}

func Vote(userID, articleID int64) error {
	// check if timeout
	article := fmt.Sprintf("article:%d", articleID)
	if timeout(articleID) {
		return ErrTimeOut
	}
	// vote
	ok, err := redisclient.Client.SAdd(fmt.Sprintf("%s:%d", VotedPrefix, articleID), userID).Result()
	if err != nil {
		return err
	}
	if ok == 0 { // voted before
		return nil
	}

	// 判断这个用户是否对这篇文章投了反对票
	against := redisclient.Client.SIsMember(fmt.Sprintf("%s:%d", AgainstVotePrefix, articleID), userID).Val()
	if against {
		ok, err := cancelAgainst(userID, articleID)
		if err != nil {
			return errors.Wrap(err, "cancel against")
		}
		if !ok {
			return errors.Errorf("cancel against error")
		}
	}

	// add score
	_, err = redisclient.Client.ZIncrBy(ScoreZSetK, VoteIncrScore, article).Result()
	if err != nil {
		return err
	}

	// incr total count
	_, err = redisclient.Client.HIncrBy(VotedCountHSetK, article, 1).Result()
	if err != nil {
		panic(err)
	}
	return nil
}

func cancelAgainst(userID, articleID int64) (bool, error) {
	// 1. 从 against_user 中移除该 user_id
	removed, err := redisclient.Client.SRem(fmt.Sprintf("%s:%d", AgainstVotePrefix, articleID), userID).Result()
	if err != nil {
		return false, errors.Wrap(err, "remove from against_voted user set")
	}
	if removed != 1 {
		return false, nil
	}

	// 2. 将对该文章投反对票的计数 -1
	_, err = redisclient.Client.HIncrBy(AgainstVoteCountHSetK, fmt.Sprintf("article:%d", articleID), -1).Result()
	if err != nil {
		return false, errors.Wrap(err, "reduce from against_count")
	}

	// 3. 将该文章的分数重新 + 432
	_, err = redisclient.Client.ZIncrBy(ScoreZSetK, VoteIncrScore, strconv.FormatInt(articleID, 10)).Result()
	if err != nil {
		return false, errors.Wrap(err, "add score")
	}
	return true, nil
}

func Publish(userID int64, title, link string) error {
	// get article id
	id, err := redisclient.Client.Incr(ArticleCountK).Result()
	if err != nil {
		return err
	}

	// add voted user
	votedUserK := fmt.Sprintf("%s:%d", VotedPrefix, id)
	_, err = redisclient.Client.SAdd(votedUserK, userID).Result()
	if err != nil {
		return err
	}

	_, err = redisclient.Client.Expire(votedUserK, time.Second*OneWeekInSeconds).Result()
	if err != nil {
		return err
	}

	// store article info
	articleKey := fmt.Sprintf("article:%d", id)
	_, err = redisclient.Client.HMSet(articleKey, "title", title, "link", link).Result()
	if err != nil {
		return err
	}

	// add to score zset
	now := time.Now().UTC().Unix()
	redisclient.Client.ZAdd(ScoreZSetK, &redis.Z{
		Score:  float64(now + VoteIncrScore),
		Member: articleKey,
	})
	// add to time zset
	redisclient.Client.ZAdd(TimeZSetK, &redis.Z{
		Score:  float64(now),
		Member: articleKey,
	})

	// 文章总投票数 + 1
	_, err = redisclient.Client.HIncrBy(VotedCountHSetK, articleKey, 1).Result()
	if err != nil {
		panic(err)
	}
	return nil
}

func GetTopArticles(page, pageSize int, zSetKey string) ([]*Article, error) {
	start := (page - 1) * pageSize
	end := start + pageSize - 1
	//zrevrange score 0 2 withscores
	zs, err := redisclient.Client.ZRevRangeWithScores(zSetKey, int64(start), int64(end)).Result()
	if err != nil {
		panic(err)
	}

	articleMap := map[int64]*Article{}
	ids := []int64{}
	for _, z := range zs {
		id, _ := strconv.ParseInt(z.Member.(string), 10, 64)
		score := z.Score
		articleMap[id] = &Article{
			ID:    id,
			Score: int64(score),
		}
		ids = append(ids, id)
	}

	// get post time of all articles
	for _, id := range ids {
		articleKey := fmt.Sprintf("article:%d", id)
		t, err := redisclient.Client.ZScore(TimeZSetK, articleKey).Result()
		if err != nil {
			if err == redis.Nil {
				log.Printf("error: %d is not in time zset ", id)
			}
			continue
		}
		articleMap[id].PublishTime = time.Unix(int64(t), 0).UTC()
	}

	// get title and link
	for _, id := range ids {
		articleKey := fmt.Sprintf("article:%d", id)
		m, err := redisclient.Client.HGetAll(articleKey).Result()
		if err != nil {
			if err == redis.Nil {
				log.Printf("error: no article:%d hash set", id)
			}
			continue
		}
		articleMap[id].Title = m["title"]
		articleMap[id].Link = m["link"]
	}
	var articles []*Article
	for _, id := range ids {
		articles = append(articles, articleMap[id])
	}
	return articles, nil
}

func AddRemoveGroups(articleID int64, toAddGroupName, toRemGroupName []string) {
	for _, g := range toAddGroupName {
		redisclient.Client.SAdd(fmt.Sprintf("%s:%s", GroupPrefix, g), articleID)
	}
	for _, g := range toRemGroupName {
		redisclient.Client.SRem(fmt.Sprintf("%s:%s", GroupPrefix, g), articleID)
	}
}

// zinterstore mzset2 2 mset mzset aggregate max
func GetGroupArticles(groupName string, page, pageSize int) ([]*Article, error) {
	// score:{groupName}
	scoreGroup := fmt.Sprintf(ScoreZSetK+":%s", groupName)
	ok, err := redisclient.Client.Exists(scoreGroup).Result()
	if err != nil {
		if err != redis.Nil {
			log.Println(err)
		}
	}

	// score & group:{groupName}
	if ok == 0 {
		count, err := redisclient.Client.ZInterStore(scoreGroup, &redis.ZStore{
			Keys: []string{ScoreZSetK, fmt.Sprintf(GroupPrefix+":%s", groupName)},
			//Weights:   []float64{1, 0},
			Aggregate: "MAX",
		}).Result()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("interstore count:", count)
		redisclient.Client.Expire(scoreGroup, 60*time.Second)
	}
	article, err := GetTopArticles(page, pageSize, scoreGroup)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return article, nil
}

func AgainstVote(userID, articleID int64) error {
	// check if timeout
	if timeout(articleID) {
		return ErrTimeOut
	}

	// 把 user_id 加入到 反对投票的用户集合中 against_voted:{articleID}
	added := redisclient.Client.SAdd(fmt.Sprintf("%s:%d", AgainstVotePrefix, articleID), userID).Val()
	if added == 0 {
		// against before
		return nil
	}

	// 判断这个文章是否之前被该用户投了 赞成票
	exist, err := redisclient.Client.SIsMember(fmt.Sprintf("%s:%d", VotedPrefix, articleID), userID).Result()
	if err != nil {
		return err
	}
	if exist {
		ok, err := cancelVote(userID, articleID)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("cancel vote error")
		}
	}

	// 扣掉对应的分
	_, err = redisclient.Client.ZIncrBy(ScoreZSetK, -VoteIncrScore, strconv.FormatInt(articleID, 10)).Result()
	if err != nil {
		return err
	}

	// 统计该文章的反对票总数
	_ = redisclient.Client.HIncrBy(AgainstVoteCountHSetK, fmt.Sprintf("article:%d", articleID), 1).Val()
	return nil
}

func cancelVote(userID, articleID int64) (bool, error) {
	// remove voted user id set
	_, err := redisclient.Client.SRem(fmt.Sprintf("voted:%d", articleID), userID).Result()
	if err != nil {
		return false, err
	}
	// down score
	article := fmt.Sprintf("article:%d", articleID)
	_, err = redisclient.Client.ZIncrBy(ScoreZSetK, -VoteIncrScore, article).Result()
	if err != nil {
		return false, err
	}
	// down voted count
	_, err = redisclient.Client.HIncrBy(VotedCountHSetK, article, -1).Result()
	if err != nil {
		return false, err
	}
	return true, nil
}
