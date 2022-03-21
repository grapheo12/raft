/**
 Defines the Raft Node stuct which contains
 the necessary variables for storing the state
 of the node as well the channels for communication

 The three distinct states in which a node can exist are handled separately in their corresponding source file are:
 * Leader    :  node_leader.go
 * Canditate :  node_candidate.go
 * Follower  :  node_follower.go

 Once a RaftNode is initialized , the control flows to NodeMain() where is repeatedly polls its state and handles requests accrodingly

**/
package raft

import (
	"context"
	"raft/internal/lo"
	"raft/pkg/network"
	"raft/pkg/rpc"
	"time"
)

const (
	FOLLOWER  = iota
	CANDIDATE = iota
	LEADER    = iota
)

type RaftNode struct {
	n               *network.Network
	NUMNODES        int32 // number of nodes known to everyone
	nId             int32
	voteRequestQId  int32
	voteResponseQId int32
	logRequestQId   int32
	logResponseQId  int32

	voteRequestCh  network.Queue
	voteResponseCh network.Queue
	logRequestCh   network.Queue
	logResponseCh  network.Queue

	ClientIn  chan []byte
	ClientOut chan []byte

	State int   // FOLLOWER, CANDIDATE, LEADER
	Term  int32 // Current Election Term
	Log   LogType

	SentLen  map[int32]int32 // node -> sent len to node
	AckedLen map[int32]int32 // node -> acked len by node

	CurrLeaderId  int32          // -1 for no leader
	VotesReceived map[int32]bool // set of votes received
	VotedFor      int32          // -1 for not voted yet

	electionMinTimeout time.Duration
	electionMaxTimeout time.Duration
	voteTimer          time.Duration
	commitTimeout      time.Duration

	StopNode context.CancelFunc

	heartbeatStarted bool
	heartbeatCancel  context.CancelFunc

	voteReqSent bool
}

func (r *RaftNode) Init(
	n *network.Network,
	voteRequestQId, voteResponseQId int32,
	logRequestQId, logResponseQId int32,
	eMinT, eMaxT, cT time.Duration) {

	r.n = n
	r.nId = n.NodeId
	r.NUMNODES = int32(len(n.Peers)) + 1
	r.voteRequestQId = voteRequestQId
	r.voteResponseQId = voteResponseQId
	r.logRequestQId = logRequestQId
	r.logResponseQId = logResponseQId

	r.voteRequestCh = make(network.Queue)
	r.voteResponseCh = make(network.Queue)
	r.logRequestCh = make(network.Queue)
	r.logResponseCh = make(network.Queue)

	r.ClientIn = make(chan []byte)
	r.ClientOut = make(chan []byte)

	r.n.RegisterQueue(r.logRequestQId, r.logRequestCh)
	r.n.RegisterQueue(r.logResponseQId, r.logResponseCh)
	r.n.RegisterQueue(r.voteRequestQId, r.voteRequestCh)
	r.n.RegisterQueue(r.voteResponseQId, r.voteResponseCh)

	r.electionMaxTimeout = eMaxT
	r.electionMinTimeout = eMinT
	r.voteTimer = eMinT
	r.commitTimeout = cT

	r.State = FOLLOWER
	r.Log = LogType{}
	r.Log.Init()
	r.Term = 0

	r.SentLen = make(map[int32]int32)
	r.AckedLen = make(map[int32]int32)

	r.CurrLeaderId = -1 // no leader now
	r.VotedFor = r.nId
	r.VotesReceived = make(map[int32]bool)

	ctx, _stopNode := context.WithCancel(context.Background())
	r.StopNode = _stopNode

	r.heartbeatStarted = false

	r.voteReqSent = false
	go r.NodeMain(ctx)
}

//  Repeatedly poll the state of the node and handle accordingly
func (r *RaftNode) NodeMain(ctx context.Context) {
	for {
		c, _ := context.WithCancel(ctx)
		if r.State == FOLLOWER {
			lo.RaftInfo(r.nId, "Found state FOLLOWER")
			r.Handle_Follower(c)
		} else if r.State == CANDIDATE {
			lo.RaftInfo(r.nId, "Found state CANDIDATE")
			r.Handle_Candidate(c)
		} else {
			lo.RaftInfo(r.nId, "Found state LEADER")
			if !r.heartbeatStarted {
				cc, ccCancel := context.WithCancel(c)
				r.heartbeatCancel = ccCancel
				go r.heartbeat(cc)
				r.heartbeatStarted = true
				lo.RaftInfo(r.nId, "hearbeat started")
			}
			r.Handle_Leader(c)
		}
	}
}

// Append entries to the log of the Node
func (n *RaftNode) AppendEntries(prefixLen int32, leaderCommitLen int32, suffix []*rpc.LogEntry) {
	if len(suffix) > 0 && n.Log.Length > int(prefixLen) {
		index := n.Log.Length
		if index > int(prefixLen)+len(suffix) {
			index = int(prefixLen) + len(suffix)
		}
		index--

		if n.Log.LogArray[index].Term != suffix[index-int(prefixLen)].Term {
			n.Log.LogArray = n.Log.LogArray[:prefixLen]
			n.Log.Length = int(prefixLen)
			n.Log.LastTerm = n.Log.LogArray[prefixLen-1].Term
		}
	}

	if int(prefixLen)+len(suffix) > n.Log.Length {
		n.Log.LogArray = append(n.Log.LogArray, suffix[n.Log.Length-int(prefixLen):]...)
		n.Log.Length = len(n.Log.LogArray)
		n.Log.LastTerm = n.Log.LogArray[n.Log.Length-1].Term
	}

	if leaderCommitLen > int32(n.Log.CommitLength) {
		for i := n.Log.CommitLength; i < int(leaderCommitLen); i++ {
			n.ClientOut <- n.Log.LogArray[i].Msg
		}
		n.Log.CommitLength = int(leaderCommitLen)
	}
}
