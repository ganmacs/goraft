package node

import (
	"log"
	"math/rand"
	"net"
	"sync/atomic"
	"time"

	rp "github.com/ganmacs/protoraft/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type node struct {
	addr   string
	others []string
	term   uint64
}

func (n *node) RequestVote(ctx context.Context, req *rp.VoteRequest) (*rp.VoteResult, error) {
	return &rp.VoteResult{req.Term, false}, nil
}

func (n *node) startServer() {
	log.Printf("start node listening to %s", n.addr)

	lis, err := net.Listen("tcp", n.addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return
	}

	s := grpc.NewServer()
	rp.RegisterLeaderElectionServer(s, n)
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	log.Printf("Start Listing %s", n.addr)
}

func (n *node) genTerm() uint64 {
	return atomic.AddUint64(&n.term, 1)
}

func (n *node) request(term uint64) {
	for _, adder := range n.others {
		log.Printf("request %s -> %s", n.addr, adder)
	}
}

func (n *node) startElection() {
	t := time.NewTicker(2 * time.Second)

	func(c <-chan time.Time) {
		for {
			select {
			case <-c:
				term := n.genTerm()
				n.request(term)
			}
		}
	}(t.C)
}

func Start(addr string, others []string) {
	nd := &node{
		addr,
		others,
		0,
	}
	go nd.startServer()

	r := time.Duration(rand.Intn(1000)) * time.Millisecond
	time.Sleep(r)

	nd.startElection()
}
