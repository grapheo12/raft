package raft

import (
	"raft/pkg/rpc"
)

type LogType struct {
	LogArray     []*rpc.LogEntry
	Length       int
	CommitLength int
	LastTerm     int32
}

func (l *LogType) Init() {
	l.LogArray = []*rpc.LogEntry{}
	l.Length = 0
	l.CommitLength = 0
	l.LastTerm = 0
}

func (l *LogType) Append(e *rpc.LogEntry) {
	l.LogArray = append(l.LogArray, e)
	l.Length++
	if e.Term > l.LastTerm {
		l.LastTerm = e.Term
	}
}

func (l *LogType) AppendEntries(prefixLen int32, leaderCommitLen int32, suffix []*rpc.LogEntry) {
	if len(suffix) > 0 && l.Length > int(prefixLen) {
		index := l.Length
		if index > int(prefixLen)+len(suffix) {
			index = int(prefixLen) + len(suffix)
		}
		index--

		if l.LogArray[index].Term != suffix[index-int(prefixLen)].Term {
			l.LogArray = l.LogArray[:prefixLen]
			l.Length = int(prefixLen)
			l.LastTerm = l.LogArray[prefixLen-1].Term
		}
	}

	if int(prefixLen)+len(suffix) > l.Length {
		l.LogArray = append(l.LogArray, suffix[l.Length-int(prefixLen):]...)
		l.Length = len(l.LogArray)
		l.LastTerm = l.LogArray[l.Length-1].Term
	}

	if leaderCommitLen > int32(l.CommitLength) {
		// TODO: DELIVER
		l.CommitLength = int(leaderCommitLen)
	}
}
