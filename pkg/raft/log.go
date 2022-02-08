package raft

// PLEASE DO SOMETHING ABOUT MSG
type Msg struct {
}

type LogEntry struct {
	m    Msg
	term int
	next *LogEntry
}

type Log struct {
	head   *LogEntry
	length int
}

func (l *Log) initLog() {
	l.head = nil
	l.length = 0
}

func (l *Log) append(e *LogEntry) {
	if l.length == 0 {
		l.head = e
		l.length += 1
		return
	}

	temp := l.head
	for i := 0; i < l.length-1; i++ {
		temp = temp.next
	}
	temp.next = e
	l.length += 1

}
