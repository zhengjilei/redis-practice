package counter

import (
	"fmt"
	"testing"
	"time"
)

func TestUpdateCounter(t *testing.T) {
	for i := 0; i < 20; i++ {
		updateCounter(1)
	}
	time.Sleep(14 * time.Second)
	for i := 0; i < 9; i++ {
		updateCounter(2)
	}
	time.Sleep(60 * time.Second)
	for i := 0; i < 13; i++ {
		updateCounter(1)
	}
}

func TestGetCounter(t *testing.T) {
	ts, count, err := getCounter(60)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(ts)
	fmt.Println(count)
	fmt.Println("----------")
	ts, count, err = getCounter(3600)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(ts)
	fmt.Println(count)
	fmt.Println("----------")
}

func TestCleaner(t *testing.T) {
	cleaner()
}