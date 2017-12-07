package main

import (
	"os"

	"github.com/ganmacs/protoraft/cli"
)

func main() {
	cli.Start(os.Args)
}
