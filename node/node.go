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
	addr        string
	others      []string
	term        uint64
	state       state
	heartbeatCh chan string
}

func (n *node) RequestVote(ctx context.Context, req *rp.VoteRequest) (*rp.VoteResult, error) {
	return &rp.VoteResult{req.Term, false}, nil
}

func (n *node) startServer() {
	lis, err := net.Listen("tcp", n.addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return
	}

	s := grpc.NewServer()
	rp.RegisterLeaderElectionServer(s, n)
	reflection.Register(s)

	log.Printf("[%s] Start Listing", n.addr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (n *node) incTerm() uint64 {
	return atomic.AddUint64(&n.term, 1)
}

func (n *node) tryAsLeader(term uint64) bool {
	ch := make(chan bool)

	// TODO: actual node count sent message is different from n.others
	nodeCount := len(n.others)
	for _, addr := range n.others {
		go func(peer string, c chan bool) {
			conn, err := grpc.Dial(peer, grpc.WithInsecure())
			if err != nil {
				log.Fatalln("fail to dial to %s :%v", peer, err)
				c <- false
				return
			}
			defer conn.Close()
			cli := rp.NewLeaderElectionClient(conn)
			log.Printf("[%s] send request to %s", n.addr, peer)

			// FIX: CandidateId is invalid
			r, err := cli.RequestVote(context.Background(), &rp.VoteRequest{Term: term, CandidateId: 0})
			if err != nil {
				log.Fatalf("could not request vote: %v", err)
				c <- false
				return
			}

			if r.VoteGranted {
				c <- true
			}

		}(addr, ch)
	}

	var accepteCount = 0
	for i := 0; i < nodeCount; i++ {
		if <-ch { // blocking
			accepteCount++
		}
	}

	return nodeCount/2 < accepteCount
}

func (n *node) startElection() {
	n.state.asCandidate()
	term := n.incTerm()

	if n.tryAsLeader(term) {
		n.asLeader()
	} else {
	}
}

func (n *node) asLeader() {
	// send heartbeat
}

func (n *node) asFollower() {
	log.Printf("[%s] as Follower\n", n.addr)
	n.state.asFollower()

	t := time.NewTicker(time.Duration(rand.Intn(3000)) * time.Millisecond)

	for {

		log.Printf("[%s] Waiting heartbeat", n.addr)
		select {
		case <-t.C:
			goto START_ELECTION // break
		case addr := <-n.heartbeatCh:
			log.Printf("[%s] heatbeat recieved from %s", n.addr, addr)
		}
	}

START_ELECTION:
	n.startElection()
}

func Start(addr string, others []string) {
	nd := &node{
		addr,
		others,
		0,
		state{},
		make(chan string),
	}

	go nd.startServer()
	// everyone is follower at first
	nd.asFollower()
}
