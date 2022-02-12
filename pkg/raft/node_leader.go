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
		return errors.New("Asontyag")
	}

	return nil
}

func (n *RaftNode) commitEntries() {
	for n.Log.CommitLength < n.Log.Length {
		acks := 0
		for nodeId := 0; nodeId < int(NUMNODES); nodeId++ {
			if n.AckedLen[int32(nodeId)] > int32(n.Log.CommitLength) {
				acks++
			}
		}

		if acks >= int(math.Ceil(float64(NUMNODES)+1.0/2)) {
			// DELIVER log[commitlength].msg
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
}

func (n *RaftNode) heartbeat(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			for i := 0; i < int(NUMNODES); i++ {
				if i != int(n.nId) {
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
	case data := <-n.voteRequestCh:
		logResp := rpc.LogResponseMsg{}
		err := logResp.Unmarshal(data.Data)
		if err != nil {
			lo.RaftError(n.nId, err.Error(), data)
			return
		}

		if errr := n.resetAsLeader(logResp.FollowerTerm); errr != nil {
			return
		}

	// Ignore
	case data := <-n.voteResponseCh:
		logResp := rpc.LogResponseMsg{}
		err := logResp.Unmarshal(data.Data)
		if err != nil {
			lo.RaftError(n.nId, err.Error(), data)
			return
		}

		if errr := n.resetAsLeader(logResp.FollowerTerm); errr != nil {
			return
		}
	// Ignore
	case data := <-n.logRequestCh:
		logResp := rpc.LogResponseMsg{}
		err := logResp.Unmarshal(data.Data)
		if err != nil {
			lo.RaftError(n.nId, err.Error(), data)
			return
		}

		if errr := n.resetAsLeader(logResp.FollowerTerm); errr != nil {
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
