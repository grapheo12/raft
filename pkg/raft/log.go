// Implements the Replicated Log for the Raft Nodes
package raft

import (
	"fmt"
	"raft/pkg/rpc"
)

type LogArrayType []*rpc.LogEntry

type LogType struct {
	LogArray     LogArrayType
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

func (l LogArrayType) String() string {
	s := ""
	for _, v := range l {
		msg_string := string(v.Msg)
		msg_term := v.Term
		s += fmt.Sprintf(" < %s, %d > ", msg_string, msg_term)
	}
	return s
}
