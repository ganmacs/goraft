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
	log.Printf("[%s] Term %v, Recieve RequestVote term=%v, accept=%v", n.addr, n.state.getTerm(), req.Term, n.state.allowAsLeader(req.Term))
	return &rp.VoteResult{req.Term, n.state.allowAsLeader(req.Term)}, nil
}

func (n *node) AppendEntry(ctx context.Context, req *rp.AppendEntryRequest) (*rp.AppendEntryResult, error) {
	log.Printf("[%s] Term %v, Recieve AppendEntry RequestVote in term=%v", n.addr, n.state.getTerm(), req.Term)
	n.heartbeatCh <- "somthing"
	// update term
	return &rp.AppendEntryResult{req.Term, true}, nil
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

	log.Printf("[%s] Start Listening", n.addr)
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
				log.Printf("fail to dial to %s :%v", peer, err)
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
				log.Printf("could not request vote: %v", err)
				c <- false
				return
			}

			if r.VoteGranted {
				log.Printf("[%s] Vote request is accepted from %s", n.addr, peer)
				c <- true
			} else {
				log.Printf("[%s] Vote request is rejected from %s", n.addr, peer)
				c <- false
			}
		}(addr, ch)
	}

	var accepteCount = 0
	for _ = range n.others {
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
	} else {
		log.Printf("[%s] Become Follower", nd.addr)
		nd.state.asFollower()
	}
}

func (nd *node) waitHeartBeat() {
	t := time.NewTicker(5 * time.Second)

	select {
	case <-nd.heartbeatCh:
		log.Printf("[%s] Receive heartbeat", nd.addr)
	case <-t.C:
		log.Printf("[%s] Waiting heartbeat is timeout", nd.addr)
		nd.state.asCandidate()
	}
}

func (nd *node) sendHeartBeat() {
	term := nd.state.getTerm()
	ch := make(chan bool)

	for _, addr := range nd.others {
		go func(peer string, c chan bool) {
			conn, err := grpc.Dial(peer, grpc.WithInsecure())
			if err != nil {
				log.Printf("fail to dial to %s :%v", peer, err)
				c <- false
				return
			}
			defer conn.Close()
			cli := rp.NewLeaderElectionClient(conn)
			log.Printf("[%s] Send request to %s", nd.addr, peer)

			// FIX: Set timeout
			r, err := cli.AppendEntry(context.Background(), &rp.AppendEntryRequest{Term: term})
			if err != nil {
				log.Printf("could not request vote: %v", err)
				c <- false
				return
			}

			if r.Success {
				log.Printf("[%s] AppendEntry to %s is success", nd.addr, peer)
			} else {
				log.Printf("[%s] AppendEntry to %s is failure", nd.addr, peer)
			}
			c <- false
		}(addr, ch)
	}

	// block
	for _ = range nd.others {
		<-ch
	}
}

func (nd *node) start() {
	// To split start timing
	s := time.Duration(rand.Intn(2000)) * time.Millisecond
	time.Sleep(s)

	for {
		time.Sleep(time.Duration(1000) * time.Millisecond)

		if nd.state.isFollower() {
			nd.waitHeartBeat()
		} else if nd.state.isCandidate() {
			nd.startElection()
		} else if nd.state.isLeader() {
			nd.sendHeartBeat()
		}
	}
}

func Start(addr string, others []string) {
	rand.Seed(time.Now().UnixNano())
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
