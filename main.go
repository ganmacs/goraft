package main

import (
	"github.com/ganmacs/protoraft/starter"
)

func main() {
	addrs := []string{
		"127.0.0.1:8000",
		"127.0.0.1:8001",
		"127.0.0.1:8002",
		"127.0.0.1:8003",
	}

	starter.Start(addrs)
}
