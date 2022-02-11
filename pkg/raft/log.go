package raft

type LogEntry struct {
	Msg  []byte
	Term int
	Next *LogEntry
}

type LogType struct {
	Head         *LogEntry
	Length       int
	LastTerm     int // term of last entry of Log
	CommitLength int
}

func (l *LogType) Init() {
	l.Head = nil
	l.Length = 0
	l.CommitLength = 0
	l.LastTerm = 0
}

func (l *LogType) Append(e *LogEntry) {
	if l.Length == 0 {
		l.Head = e
		l.Length++
		return
	}

	temp := l.Head
	for i := 0; i < l.Length-1; i++ {
		temp = temp.Next
	}
	temp.Next = e
	l.Length++
	if e.Term > l.LastTerm {
		l.LastTerm = e.Term
	}
}
