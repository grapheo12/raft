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

	CurrLeaderId  int32        // -1 for no leader
	VotesReceived map[int]bool // set of votes received
	VotedFor      int32        // -1 for not voted yet

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

	r.CurrLeaderId = -1 // no leader now
	r.VotedFor = r.nId
	r.VotesReceived = make(map[int]bool)
}

// function a node calls after a node receives VoteRequestMsg
func (n *RaftNode) Handle_VoteRequest(m VoteRequestMsg) {
	if m.CandidateTerm > n.Term {
		n.Term = m.CandidateTerm
		n.State = FOLLOWER
		n.VotedFor = -1
	}

	lastTerm := 0
	if n.Log.Length > 0 {
		lastTerm = n.Log.LastTerm
	}

	// if requester log is okay to vote for
	logOk := (m.CandidateLogTerm > lastTerm) || ((m.CandidateTerm == lastTerm) && (m.CandidateLogLen > n.Log.Length))

	if (m.CandidateTerm == n.Term) && logOk && (n.VotedFor == m.CandidateId || n.VotedFor == -1) {
		n.VotedFor = m.CandidateId
		// TODO :: send VoteResponse with true granted
	} //else {
	// TODO :: send VoteResponse with false granted
	//}

}
