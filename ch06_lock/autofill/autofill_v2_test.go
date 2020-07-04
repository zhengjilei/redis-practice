package autofill

import (
	"fmt"
	"testing"
)

func TestAutoFillV2(t *testing.T) {
	res := autoFillV2("family", "abc")
	fmt.Println(res)
}

func TestInitData(t *testing.T) {
	initGroup("family", []string{
		"abc",
		"abbzz",
		"abxzdsa",
		"abbcde",
		"abcde",
		"abcz",
		"abk",
		"abca",
		"dsa",
		"bba",
		"azb",
		"acb",
	})
}

func TestName(t *testing.T) {
	a := 'z'
	fmt.Println(string(a + 1))
}
