package stats

import (
	"fmt"
	"log"
	"testing"
)

func TestUpdateStats(t *testing.T) {
	err := updateStats("ProfilePage", "AccessTime", 0.024, 20)
	log.Println(err)
	if err != nil {
		t.Error(err)
	}

}

func TestUpdateStats2(t *testing.T) {
	err := updateStats("ProfilePage", "AccessTime", 0.035, 20)
	log.Println(err)
	if err != nil {
		t.Error(err)
	}
}
func TestUpdateStats3(t *testing.T) {
	err := updateStats("ProfilePage", "AccessTime", 0.019, 20)
	log.Println(err)
	if err != nil {
		t.Error(err)
	}
}

func TestGetStats(t *testing.T) {
	res, err := getStats("ProfilePage", "AccessTime")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
}
