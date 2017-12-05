package node

import (
	"sync"
)

type stateType uint8

const (
	followerType stateType = iota
	candidateType
	leaderType
)

type state struct {
	stateType stateType
	mutex     sync.RWMutex
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
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.stateType = candidateType
}

func (n *state) asLeader() {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.stateType = leaderType
}
