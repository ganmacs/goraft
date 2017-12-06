package node

import (
	"sync"
	"sync/atomic"
)

type stateType uint8

const (
	followerType stateType = iota
	candidateType
	leaderType
)

type state struct {
	stateType    stateType
	mutex        sync.RWMutex
	term         uint64
	lastVoteTerm uint64
}

func (s *state) allowAsLeader(t uint64) bool {
	// MUTEX
	if s.term < t {
		s.lastVoteTerm = t
		return true
	} else if s.term > t {
		return false
	}

	if s.lastVoteTerm == s.term {
		return false
	}
	s.lastVoteTerm = s.term
	return true
}

func (s *state) getTerm() uint64 {
	// MUTEX
	return s.term
}

func (n *state) isFollower() bool {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	return n.stateType == followerType
}

func (n *state) isCandidate() bool {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	return n.stateType == candidateType
}

func (n *state) isLeader() bool {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	return n.stateType == leaderType
}

// TODO: use CAS
func (n *state) asFollower() {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.stateType = followerType
}

func (n *state) asCandidate() {
	atomic.AddUint64(&n.term, 1)
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.stateType = candidateType
}

func (n *state) asLeader() {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.stateType = leaderType
}
