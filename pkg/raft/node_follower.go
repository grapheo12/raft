package raft

import (
	"context"
	"errors"
	"math/rand"
	"raft/internal/lo"
	"raft/pkg/rpc"
	"time"
)

func (n *RaftNode) resetAsFollower(ct int32) error {
	if ct > n.Term {
		n.Term = ct
		n.State = FOLLOWER
		n.VotedFor = -1
		lo.RaftInfo(n.nId, "Higher term found : [", ct, "], state [FOLLOWER -> FOLLOWER]")
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
		n.voteReqSent = false
		lo.RaftInfo(n.nId, "Timeout occured, state [FOLLOWER -> CANDIDATE]")
		return
	case data := <-n.voteRequestCh:
		voteReq := rpc.VoteRequestMsg{}
		err := voteReq.Unmarshal(data.Data)
		if err != nil {
			lo.RaftError(n.nId, err.Error(), data)
			return
		}

		lo.RaftInfo(n.nId, "Received VoteRequest from ", voteReq.CandidateId)

		n.resetAsFollower(voteReq.CandidateTerm)

		lastTerm := int32(0)
		if n.Log.Length > 0 {
			lastTerm = n.Log.LastTerm
		}

		// if requester log is okay to vote for
		logOk := (voteReq.CandidateLogTerm > lastTerm) ||
			((voteReq.CandidateLogTerm == lastTerm) &&
				(int(voteReq.CandidateLogLen) >= n.Log.Length))

		lo.RaftInfo(n.nId, "log of candidate is okay?", logOk)

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
			lo.RaftInfo(n.nId, "Sent vote for", voteReq.CandidateId)
		} else {
			// send VoteResponse to cadidateId with false granted
			resp := rpc.VoteResponseMsg{
				VoterId:   n.nId,
				VoterTerm: n.Term,
				Granted:   false,
			}
			send_data, _ := resp.Marshal()
			n.n.Send(voteReq.CandidateId, n.voteResponseQId, send_data)
			lo.RaftInfo(n.nId, "Sent vote against", voteReq.CandidateId)
		}

	case data := <-n.voteResponseCh:
		voteResp := rpc.VoteResponseMsg{}
		err := voteResp.Unmarshal(data.Data)
		if err != nil {
			lo.RaftError(n.nId, err.Error(), data)
			return
		}

		lo.RaftInfo(n.nId, "Recieved VoteResponse from", voteResp.VoterId)

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

		lo.RaftInfo(n.nId, "Received LogRequest from", logReq.LeaderId)
		n.resetAsFollower(logReq.LeaderTerm)

		if n.Term == logReq.LeaderTerm {
			n.CurrLeaderId = logReq.LeaderId
		}

		logOk := (n.Log.Length >= int(logReq.PrefixLen)) &&
			((logReq.PrefixLen == 0) ||
				(n.Log.LogArray[logReq.PrefixLen-1].Term == logReq.PrefixTerm))

		if (logReq.LeaderTerm == n.Term) && logOk {
			n.AppendEntries(logReq.PrefixLen, logReq.CommitLen, logReq.Suffix)
			ack := int(logReq.PrefixLen) + len(logReq.Suffix)

			resp := rpc.LogResponseMsg{
				FollowerId:       n.nId,
				FollowerTerm:     n.Term,
				LogCommitAck:     int32(ack),
				LogCommitSuccess: true,
			}
			send_data, _ := resp.Marshal()

			n.n.Send(logReq.LeaderId, n.logResponseQId, send_data)
			lo.RaftInfo(n.nId, "Sent LogResponse to", logReq.LeaderId, "commit: true")

		} else {
			resp := rpc.LogResponseMsg{
				FollowerId:       n.nId,
				FollowerTerm:     n.Term,
				LogCommitAck:     int32(0),
				LogCommitSuccess: false,
			}
			send_data, _ := resp.Marshal()

			n.n.Send(logReq.LeaderId, n.logResponseQId, send_data)
			lo.RaftInfo(n.nId, "Sent LogResponse to", logReq.LeaderId, "commit: false")
		}

	case data := <-n.logResponseCh:
		logResp := rpc.LogResponseMsg{}
		err := logResp.Unmarshal(data.Data)
		if err != nil {
			lo.RaftError(n.nId, err.Error(), data)
			return
		}

		lo.RaftInfo(n.nId, "Received LogResponse from", logResp.FollowerId)

		if errr := n.resetAsFollower(logResp.FollowerTerm); errr != nil {
			return
		}
		// Ignore

	}
}
