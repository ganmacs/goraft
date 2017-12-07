package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ganmacs/protoraft/node"
)

type CommandLineOption struct {
	ListenAddr string
	Peers      string
}

func Start(args []string) {
	opt := CommandLineOption{}
	fl := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fl.StringVar(&opt.ListenAddr, "addr", "127.0.0.1:8080", "Listeing addr")
	fl.StringVar(&opt.Peers, "peers", "", "peers")

	err := fl.Parse(args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	node.Start(opt.ListenAddr, strings.Split(opt.Peers, ","))
}
