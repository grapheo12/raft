package raft

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"raft/internal/lo"
	"raft/pkg/network"
	"raft/pkg/rpc"
	"time"
)

var NUMNODES int32 // number of nodes known to everyone

const (
	FOLLOWER  = iota
	CANDIDATE = iota
	LEADER    = iota
)

type RaftNode struct {
	n               *network.Network
	nId             int32
	voteRequestQId  int32
	voteResponseQId int32
	logRequestQId   int32
	logResponseQId  int32

	voteRequestCh  network.Queue
	voteResponseCh network.Queue
	logRequestCh   network.Queue
	logResponseCh  network.Queue

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
	commitTimeout      time.Duration

	StopNode context.CancelFunc
}

func (r *RaftNode) Init(
	n *network.Network,
	voteRequestQId, voteResponseQId int32,
	logRequestQId, logResponseQId int32,
	eMinT, eMaxT, cT time.Duration) {

	r.n = n
	r.nId = n.NodeId
	r.voteRequestQId = voteRequestQId
	r.voteResponseQId = voteResponseQId
	r.logRequestQId = logRequestQId
	r.logResponseQId = logResponseQId

	r.voteRequestCh = make(network.Queue)
	r.voteResponseCh = make(network.Queue)
	r.logRequestCh = make(network.Queue)
	r.logResponseCh = make(network.Queue)

	r.n.RegisterQueue(r.logRequestQId, r.logRequestCh)
	r.n.RegisterQueue(r.logResponseQId, r.logResponseCh)
	r.n.RegisterQueue(r.voteRequestQId, r.voteRequestCh)
	r.n.RegisterQueue(r.voteResponseQId, r.voteResponseCh)

	r.electionMaxTimeout = eMaxT
	r.electionMinTimeout = eMinT
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
	go r.NodeMain(ctx)
}

func (r *RaftNode) NodeMain(ctx context.Context) {
	for {
		c, _ := context.WithCancel(ctx)
		if r.State == FOLLOWER {
			r.Handle_Follower(c)
		} else if r.State == CANDIDATE {
			r.Handle_Candidate(c)
		} else {
			r.Handle_Leader(c)
		}
	}
}

func (n *RaftNode) resetAsFollower(ct int32) error {
	if ct > n.Term {
		n.Term = ct
		n.State = FOLLOWER
		n.VotedFor = -1
		return errors.New("Podotyag")
	}

	return nil
}

func (n *RaftNode) Handle_Follower(ctx context.Context) {
	tv := n.electionMinTimeout.Milliseconds()
	tv += rand.Int63n(n.electionMaxTimeout.Milliseconds() - n.electionMinTimeout.Milliseconds())
	timeout, cancel := context.WithTimeout(ctx, time.Duration(tv*int64(time.Millisecond)))
	defer cancel()

	select {
	case <-ctx.Done():
		return
	case <-timeout.Done():
		n.State = CANDIDATE
		return
	case data := <-n.voteRequestCh:
		voteReq := rpc.VoteRequestMsg{}
		err := voteReq.Unmarshal(data.Data)
		if err != nil {
			lo.RaftError(n.nId, err.Error(), data)
			return
		}
		if errr := n.resetAsFollower(voteReq.CandidateTerm); errr != nil {
			return
		}

		var lastTerm int32 = 0
		if n.Log.Length > 0 {
			lastTerm = n.Log.LastTerm
		}

		// if requester log is okay to vote for
		logOk := (voteReq.CandidateLogTerm > lastTerm) ||
			((voteReq.CandidateTerm == lastTerm) &&
				(int(voteReq.CandidateLogLen) >= n.Log.Length))

		if (voteReq.CandidateTerm == n.Term) && logOk &&
			(n.VotedFor == voteReq.CandidateId || n.VotedFor == -1) {
			n.VotedFor = voteReq.CandidateId

			// send VoteResponse with true granted
			resp := rpc.VoteResponseMsg{
				VoterId:   n.nId,
				VoterTerm: n.Term,
				Granted:   true,
			}
			send_data, _ := resp.Marshal()
			n.n.Send(voteReq.CandidateId, n.voteResponseQId, send_data)
		} else {
			// send VoteResponse to cadidateId with false granted
			resp := rpc.VoteResponseMsg{
				VoterId:   n.nId,
				VoterTerm: n.Term,
				Granted:   false,
			}
			send_data, _ := resp.Marshal()
			n.n.Send(voteReq.CandidateId, n.voteResponseQId, send_data)
		}
	case data := <-n.voteResponseCh:
		voteResp := rpc.VoteResponseMsg{}
		err := voteResp.Unmarshal(data.Data)
		if err != nil {
			lo.RaftError(n.nId, err.Error(), data)
			return
		}
		if errr := n.resetAsFollower(voteResp.VoterTerm); errr != nil {
			return
		}
		// Ignore
	case data := <-n.logRequestCh:
		logReq := rpc.LogRequestMsg{}
		err := logReq.Unmarshal(data.Data)
		if err != nil {
			lo.RaftError(n.nId, err.Error(), data)
			return
		}
		n.resetAsFollower(logReq.LeaderTerm)

		if n.Term == logReq.LeaderTerm {
			n.CurrLeaderId = logReq.LeaderId
		}

		logOk := (n.Log.Length >= int(logReq.PrefixLen)) &&
			((logReq.PrefixLen == 0) ||
				(n.Log.LogArray[logReq.PrefixLen-1].Term == logReq.PrefixTerm))

		if (logReq.LeaderTerm == n.Term) && logOk {
			n.Log.AppendEntries(logReq.PrefixLen, logReq.CommitLen, logReq.Suffix)
			ack := int(logReq.PrefixLen) + len(logReq.Suffix)

			resp := rpc.LogResponseMsg{
				FollowerId:       n.nId,
				FollowerTerm:     n.Term,
				LogCommitAck:     int32(ack),
				LogCommitSuccess: true,
			}
			send_data, _ := resp.Marshal()

			n.n.Send(logReq.LeaderId, n.logResponseQId, send_data)
		} else {
			resp := rpc.LogResponseMsg{
				FollowerId:       n.nId,
				FollowerTerm:     n.Term,
				LogCommitAck:     int32(0),
				LogCommitSuccess: false,
			}
			send_data, _ := resp.Marshal()

			n.n.Send(n.nId, n.logResponseQId, send_data)
		}

	case data := <-n.logResponseCh:
		logResp := rpc.LogResponseMsg{}
		err := logResp.Unmarshal(data.Data)
		if err != nil {
			lo.RaftError(n.nId, err.Error(), data)
			return
		}
		if errr := n.resetAsFollower(logResp.FollowerTerm); errr != nil {
			return
		}
		// Ignore

	}
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
