package node

import (
	"log"
	"math/rand"
	"net"
	"time"

	rp "github.com/ganmacs/protoraft/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type node struct {
	addr        string
	others      []string
	state       state
	heartbeatCh chan string
}

func (n *node) RequestVote(ctx context.Context, req *rp.VoteRequest) (*rp.VoteResult, error) {
	log.Printf("[%s] Recieve RequestVote from --", n.addr)
	return &rp.VoteResult{req.Term, n.state.allowAsLeader(req.Term)}, nil
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

func (n *node) tryAsLeader() bool {
	term := n.state.getTerm()
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
			log.Printf("[%s] Send request to %s", n.addr, peer)

			// FIX: CandidateId is invalid
			// FIX: Set timeout
			r, err := cli.RequestVote(context.Background(), &rp.VoteRequest{Term: term, CandidateId: 0})
			if err != nil {
				log.Fatalf("could not request vote: %v", err)
				c <- false
				return
			}

			if r.VoteGranted {
				log.Printf("[%s] Vote request is accepted from %s", n.addr, peer)
				c <- true
			} else {
				log.Printf("[%s] Vote request is rejected from %s", n.addr, peer)
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

func (nd *node) startElection() {
	if nd.tryAsLeader() {
		log.Printf("[%s] Become Leader", nd.addr)
		nd.state.asLeader()
	}
}

func (nd *node) start() {
	t := time.NewTicker(time.Duration(rand.Intn(4000)) * time.Millisecond)

	for {
		select {
		case <-t.C:
			if nd.state.isFollower() {
			} else if nd.state.isCandidate() {
				go nd.startElection()
			} else if nd.state.isLeader() {
			}
		case addr := <-nd.heartbeatCh:
			log.Printf("[%s] heatbeat recieved from %s", nd.addr, addr)
		}
	}
}

func Start(addr string, others []string) {
	nd := &node{
		addr,
		others,
		state{},
		make(chan string),
	}

	go nd.startServer()

	// everyone is candidate at first
	log.Printf("[%s] Become Candidate", nd.addr)
	nd.state.asCandidate()
	nd.start()
}
