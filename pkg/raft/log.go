package raft

import (
	"raft/pkg/rpc"
)

type LogType struct {
	LogArray     []rpc.LogEntry
	Length       int
	CommitLength int
	LastTerm     int32
}

func (l *LogType) Init() {
	l.LogArray = []rpc.LogEntry{}
	l.Length = 0
	l.CommitLength = 0
	l.LastTerm = 0
}

func (l *LogType) Append(e rpc.LogEntry) {
	l.LogArray = append(l.LogArray, e)
	l.Length++
	if e.Term > l.LastTerm {
		l.LastTerm = e.Term
	}
}
