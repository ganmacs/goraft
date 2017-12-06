package starter

import (
	"github.com/ganmacs/protoraft/node"
)

func Start(addrs []string) {
	ch := make(chan int)

	for i := 0; i < len(addrs); i++ {
		var tmp []string
		// to avoid override addr
		tmp = append(tmp, addrs[:i]...)
		tmp = append(tmp, addrs[i+1:]...)
		go node.Start(addrs[i], tmp)
	}

	<-ch
}
