package raft

import (
	"math"
	"raft/pkg/network"
	"raft/pkg/rpc"
	"time"
)

var NUMNODES int // number of nodes known to everyone

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

	State int   // FOLLOWER, CANDIDATE, LEADER
	Term  int32 // Current Election Term
	Log   LogType

	CurrLeaderId  int32          // -1 for no leader
	VotesReceived map[int32]bool // set of votes received
	VotedFor      int32          // -1 for not voted yet

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
	r.VotesReceived = make(map[int32]bool)
}

// function a node calls after a node receives VoteRequestMsg
func (n *RaftNode) Handle_VoteRequest(m rpc.VoteRequestMsg) {
	if m.CandidateTerm > n.Term {
		n.Term = m.CandidateTerm
		n.State = FOLLOWER
		n.VotedFor = -1
	}

	var lastTerm int32 = 0
	if n.Log.Length > 0 {
		lastTerm = n.Log.LastTerm
	}

	// if requester log is okay to vote for
	logOk := (m.CandidateLogTerm > lastTerm) || ((m.CandidateTerm == lastTerm) && (int(m.CandidateLogLen) > n.Log.Length))

	if (m.CandidateTerm == n.Term) && logOk && (n.VotedFor == m.CandidateId || n.VotedFor == -1) {
		n.VotedFor = m.CandidateId
		// TODO :: send VoteResponse with true granted
	} //else {
	// TODO :: send VoteResponse to cadidateId with false granted
	//}
}

// a node counts its votes
func (n *RaftNode) countVotes() int {
	sum := 0
	for _, granted := range n.VotesReceived {
		if granted {
			sum++
		}
	}
	return sum
}

// function a node will invoke after receiving a VoteResponseMsg
func (n *RaftNode) Handle_VoteResponse(m rpc.VoteResponseMsg) {
	if n.State == CANDIDATE && n.Term == m.VoterTerm && m.Granted {
		n.VotesReceived[m.VoterId] = true
		if n.countVotes() > int(math.Ceil((float64(NUMNODES)+1)/2)) {
			n.State = LEADER
			n.CurrLeaderId = n.nId
			// TODO ::
			// cancel election timer
			// start REPLICATELOG
		}
	} else if m.VoterTerm > n.Term {
		n.Term = m.VoterTerm
		n.State = FOLLOWER
		n.VotedFor = -1
		// TODO ::
		// cancel election timer
	}
}
