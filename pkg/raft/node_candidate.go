/**
 Source file for handling the candidate state
**/
package raft

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"raft/internal/lo"
	"raft/pkg/rpc"
	"time"
)

func (n *RaftNode) resetAsCandidate(ct int32) error {
	if ct > n.Term {
		n.Term = ct
		n.State = FOLLOWER
		n.VotedFor = -1
		n.voteReqSent = false
		lo.RaftInfo(n.nId, "Higher term found : [", ct, "], state [CANDIDATE -> FOLLOWER]")
		return errors.New("Protyahar")
	}
	return nil
}

func (n *RaftNode) isMajority() bool {
	numVotes := len(n.VotesReceived)
	lo.RaftInfo(n.nId, "votes received =", numVotes)
	return (numVotes >= int(math.Ceil((float64(n.NUMNODES)+1.0)/2)))
}

func (n *RaftNode) Handle_Candidate(ctx context.Context) {
	if !n.voteReqSent {
		// Prepare to send a VoteRequest
		n.Term++
		n.VotesReceived = make(map[int32]bool)
		n.VotesReceived[n.nId] = true
		n.VotedFor = n.nId

		voteReq := rpc.VoteRequestMsg{
			CandidateId:      n.nId,
			CandidateTerm:    n.Term,
			CandidateLogLen:  int32(n.Log.Length),
			CandidateLogTerm: n.Log.LastTerm,
		}
		send_data, _ := voteReq.Marshal()
		n.n.Broadcast(n.voteRequestQId, send_data)
		lo.RaftInfo(n.nId, "Broadcasted VoteRequest")
		n.voteReqSent = true
	}
	// Generate a timeout value in range [n.electionMinTimeout, n.electionMaxTimeout]
	tv := n.electionMinTimeout.Milliseconds()
	tv += rand.Int63n(n.electionMaxTimeout.Milliseconds() - n.electionMinTimeout.Milliseconds())
	timeout, cancel := context.WithTimeout(ctx, time.Duration(tv*int64(time.Millisecond)))
	defer cancel()

	select {
	case <-ctx.Done():
		return
	case <-timeout.Done():
		// Start voting again
		n.Term--
		n.voteReqSent = false
		lo.RaftInfo(n.nId, "Restarting vote")
	case data := <-n.voteRequestCh:
		// Received vote request from another node
		voteReq := rpc.VoteRequestMsg{}
		err := voteReq.Unmarshal(data.Data)
		if err != nil {
			lo.RaftError(n.nId, err.Error(), data)
			return
		}

		lo.RaftInfo(n.nId, "Received VoteRequest from", voteReq.CandidateId)

		if errr := n.resetAsCandidate(voteReq.CandidateTerm); errr != nil {
			return
		}
		// Ignore
	case data := <-n.voteResponseCh:
		// Received a vote
		voteResp := rpc.VoteResponseMsg{}
		err := voteResp.Unmarshal(data.Data)
		if err != nil {
			lo.RaftError(n.nId, err.Error(), data)
			return
		}

		lo.RaftInfo(n.nId, "Received vote response from", voteResp.VoterId, "voted", voteResp.Granted)

		if errr := n.resetAsCandidate(voteResp.VoterTerm); errr != nil {
			return
		}

		if voteResp.Granted {
			n.VotesReceived[voteResp.VoterId] = true
		}

		// Check if quorum has been achieved
		if n.isMajority() {
			n.State = LEADER
			n.voteReqSent = false

			for i := int32(0); i < n.NUMNODES; i++ {
				if i != n.nId {
					n.SentLen[i] = int32(n.Log.Length)
					n.AckedLen[i] = 0
				}
			}

			lo.RaftInfo(n.nId, "Achieved quorum, [CANDIDATE -> LEADER]")
		}

	case data := <-n.logRequestCh:
		logReq := rpc.LogRequestMsg{}
		err := logReq.Unmarshal(data.Data)
		if err != nil {
			lo.RaftError(n.nId, err.Error(), data)
			return
		}

		lo.RaftInfo(n.nId, "Received LogRequest from", logReq.LeaderId)

		if errr := n.resetAsCandidate(logReq.LeaderTerm); errr != nil {
			return
		}
		// Ignore

	case data := <-n.logResponseCh:
		logResp := rpc.LogResponseMsg{}
		err := logResp.Unmarshal(data.Data)
		if err != nil {
			lo.RaftError(n.nId, err.Error(), data)
			return
		}

		lo.RaftInfo(n.nId, "Received LogResponse from", logResp.FollowerId)

		if errr := n.resetAsCandidate(logResp.FollowerTerm); errr != nil {
			return
		}
		// Ignore
	}
}
