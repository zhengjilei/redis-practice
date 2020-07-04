package ch01

import (
	"encoding/json"
	"fmt"
	"github.com/zhengjilei/redis-practice/common/redisclient"
	"testing"
)

func TestPublish(t *testing.T) {

	userID1 := 1
	userID2 := 2
	userID3 := 3

	err := Publish(int64(userID1), "title1", "link1")
	if err != nil {
		t.Error(err)
	}

	err = Publish(int64(userID2), "title2", "link2")
	if err != nil {
		t.Error(err)
	}

	err = Publish(int64(userID3), "title3", "link3")
	if err != nil {
		t.Error(err)
	}
}

func TestVote(t *testing.T) {
	// user1 vote for article3
	err := Vote(1, 3)
	if err != nil {
		t.Error(err)
	}

	// user8 vote for article2
	err = Vote(8, 2)
	if err != nil {
		t.Error(err)
	}
	// user5 vote for article2
	err = Vote(5, 2)
	if err != nil {
		t.Error(err)
	}
}

func TestPublish2(t *testing.T) {
	err := Publish(10, "title10", "link10")
	if err != nil {
		t.Error(err)
	}
}
func TestRedis(t *testing.T) {
	val, err := redisclient.Client.Get("a").Result()
	fmt.Println(err)
	fmt.Println(val + " <= <val>")

	a := A{
		Name: "ethan",
		Age:  21,
	}
	fmt.Println("-----------")
	res, err := redisclient.Client.Set("aaa", a, -1).Result()

	fmt.Println(res)
	fmt.Println(err)
}

func TestRedis2(t *testing.T) {
	val, err := redisclient.Client.Get("aaa").Result()
	fmt.Println(err)
	fmt.Println(val + " <= <val>")

	bys := []byte(val)
	var a A
	err = json.Unmarshal(bys, &a)
	if err != nil {
		panic(err)
	}
	fmt.Println(a)
}

type A struct {
	Name string `json:"name"`
	Age  int64  `json:"age"`
}

func (a A) MarshalBinary() (data []byte, err error) {
	return json.Marshal(a)
}

func TestGetTopArticles(t *testing.T) {
	articles, err := GetTopArticles(1, 10, "score")
	if err != nil {
		t.Error(err)
	}
	for _, v := range articles {
		fmt.Printf("%+v\n", *v)
	}
}
func TestAddRemoveGroups(t *testing.T) {
	AddRemoveGroups(1, []string{"g1", "g3"}, []string{})
	AddRemoveGroups(2, []string{"g3", "g4", "g5"}, []string{})
	AddRemoveGroups(3, []string{"g3", "g5", "g6"}, []string{})
}

func TestGetGroupArticles(t *testing.T) {
	articles, err := GetGroupArticles("g5", 1, 10)

	if err != nil {
		t.Error(err)
	}
	for _, v := range articles {
		fmt.Printf("%+v\n", *v)
	}
}

func TestAgainstVote(t *testing.T) {
	AgainstVote(2, 1)
}
func TestAgainstVote2(t *testing.T) {
	AgainstVote(3, 2)
}
func TestAgainstVote3(t *testing.T) {
	AgainstVote(2, 2)
}
