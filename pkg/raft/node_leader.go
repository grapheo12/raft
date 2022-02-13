package raft

import (
	"context"
	"errors"
	"math"
	"raft/internal/lo"
	"raft/pkg/rpc"
	"time"
)

func (n *RaftNode) resetAsLeader(ct int32) error {
	if ct > n.Term {
		n.Term = ct
		n.State = FOLLOWER
		n.VotedFor = -1
		n.heartbeatCancel()
		n.heartbeatStarted = false
		lo.RaftInfo(n.nId, "Higher term found : [", ct, "], state [LEADER -> FOLLOWER]")
		return errors.New("Asontyag")
	}

	return nil
}

func (n *RaftNode) commitEntries() {
	for n.Log.CommitLength < n.Log.Length {
		acks := 0
		for nodeId := 0; nodeId < int(n.NUMNODES); nodeId++ {
			if n.AckedLen[int32(nodeId)] > int32(n.Log.CommitLength) {
				acks++
			}
		}

		if acks >= int(math.Ceil((float64(n.NUMNODES)+1.0)/2)) {
			n.ClientOut <- n.Log.LogArray[n.Log.CommitLength].Msg
			n.Log.CommitLength++
		} else {
			break
		}
	}
}

func (n *RaftNode) replicateLog(followerId int32) {
	prefixLen := n.SentLen[followerId]
	suffix := n.Log.LogArray[prefixLen:]
	prefixTerm := int32(0)
	if prefixLen > 0 {
		prefixTerm = n.Log.LogArray[prefixLen-1].Term
	}

	resp := rpc.LogRequestMsg{
		LeaderId:   n.nId,
		LeaderTerm: n.Term,
		PrefixLen:  prefixLen,
		PrefixTerm: prefixTerm,
		CommitLen:  int32(n.Log.CommitLength),
		Suffix:     suffix,
	}

	send_data, _ := resp.Marshal()

	n.n.Send(followerId, n.logRequestQId, send_data)
	lo.RaftInfo(n.nId, "Sent LogRequest", followerId)
}

func (n *RaftNode) heartbeat(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			for i := 0; i < int(n.NUMNODES); i++ {
				if i != int(n.nId) {
					// lo.RaftInfo(n.nId, "i= ", i, "NUMNODES= ", n.NUMNODES)
					n.replicateLog(int32(i))
				}
			}
			time.Sleep(n.electionMinTimeout / 2)
		}
	}
}

func (n *RaftNode) Handle_Leader(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case data := <-n.ClientIn:
		n.Log.Append(&rpc.LogEntry{
			Msg:  data,
			Term: n.Term,
		})
		for i := 0; i < int(n.NUMNODES); i++ {
			if i != int(n.nId) {
				n.replicateLog(int32(i))
			}
		}
	case data := <-n.voteRequestCh:
		voteReq := rpc.VoteRequestMsg{}
		err := voteReq.Unmarshal(data.Data)
		if err != nil {
			lo.RaftError(n.nId, err.Error(), data)
			return
		}

		lo.RaftInfo(n.nId, "Received VoteRequest from", voteReq.CandidateId)

		if errr := n.resetAsLeader(voteReq.CandidateTerm); errr != nil {
			return
		}

	// Ignore
	case data := <-n.voteResponseCh:
		voteResp := rpc.VoteResponseMsg{}
		err := voteResp.Unmarshal(data.Data)
		if err != nil {
			lo.RaftError(n.nId, err.Error(), data)
			return
		}

		lo.RaftInfo(n.nId, "Received VoteResponse from", voteResp.VoterId)

		if errr := n.resetAsLeader(voteResp.VoterTerm); errr != nil {
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

		if errr := n.resetAsLeader(logReq.LeaderTerm); errr != nil {
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

		if errr := n.resetAsLeader(logResp.FollowerTerm); errr != nil {
			return
		}

		if (logResp.FollowerTerm == n.Term) && n.State == LEADER {
			if logResp.LogCommitSuccess &&
				(logResp.LogCommitAck >= int32(n.AckedLen[logResp.FollowerId])) {
				n.SentLen[logResp.FollowerId] = logResp.LogCommitAck
				n.AckedLen[logResp.FollowerId] = logResp.LogCommitAck
				n.commitEntries()
			} else if n.SentLen[logResp.FollowerId] > 0 {
				n.SentLen[logResp.FollowerId]--
				n.replicateLog(logResp.FollowerId)
			}
		}

	}
}
