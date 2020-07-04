package config

import (
	"fmt"
	"testing"
)

const (
	jsonServer = `
	{
		"host":"111.111.111.111",
		"port":322,
		"max_conn":32
	}
`
)

func TestRegisterConfig(t *testing.T) {
	registerConfig("redis", "server", jsonServer)
}

func TestGetConfig(t *testing.T) {
	res, err := getConfig("redis", "server", 10)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res)
}

func TestChar(t *testing.T) {
	fmt.Println('{') // 123
	fmt.Println('z') // 122
	fmt.Println('`') // 96
	fmt.Println('a') // 97
}
