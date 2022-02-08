package raft

import (
	"raft/pkg/network"
	"time"
)

const (
	FOLLOWER  = iota
	CANDIDATE = iota
	LEADER    = iota
)

type RaftNode struct {
	n           *network.Network
	nId         int32
	normalQId   int32
	electionQId int32

	normalCh   network.Queue
	electionCh network.Queue

	State int // FOLLOWER, CANDIDATE, LEADER
	Term  int // Current Election Term
	Log   LogType

	electionMinTimeout time.Duration
	electionMaxTimeout time.Duration
	commitTimeout      time.Duration
}

func (r *RaftNode) Init(
	n *network.Network, normalQId, electionQId int32,
	eMinT, eMaxT, cT time.Duration) {

	r.n = n
	r.nId = n.NodeId
	r.normalQId = normalQId
	r.electionQId = electionQId

	r.normalCh = make(network.Queue)
	r.electionCh = make(network.Queue)

	r.n.RegisterQueue(r.normalQId, r.normalCh)
	r.n.RegisterQueue(r.electionQId, r.electionCh)

	r.electionMaxTimeout = eMaxT
	r.electionMinTimeout = eMinT
	r.commitTimeout = cT

	r.State = FOLLOWER
	r.Log = LogType{}
	r.Log.Init()
	r.Term = 0
}
