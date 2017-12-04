package starter

import (
	"github.com/ganmacs/protoraft/node"
)

func Start(addrs []string) {
	ch := make(chan int)

	for i := 0; i < len(addrs); i++ {
		tmp := []string{}
		copy(tmp, addrs[:i]) // to avoid override addr
		go node.Start(addrs[i], append(tmp, addrs[i+1:]...))
	}

	<-ch
}
