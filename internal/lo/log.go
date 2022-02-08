package lo

import (
	"fmt"
	"io"
	"sync"
)

type SingleLog struct {
	Sink  io.Writer
	Level int
	Lck   sync.Mutex
}

type Logger struct {
	sinks []*SingleLog
}

func (l *Logger) AddSink(w io.Writer, lvl int) {
	l.sinks = append(l.sinks, &SingleLog{
		Sink:  w,
		Level: lvl,
	})
}

func (s *SingleLog) LogLn(lvl int, v ...interface{}) {
	if lvl < s.Level {
		return
	}

	s.Lck.Lock()
	defer s.Lck.Unlock()

	fmt.Fprintln(s.Sink, v...)
}

func (l *Logger) LogLn(lvl int, n int32, module string, mode string, v ...interface{}) {
	a := make([]interface{}, 0)
	if n >= 0 {
		a = append(a, fmt.Sprintf("[Node %d][%s][%s]", n, module, mode))
	} else {
		a = append(a, fmt.Sprintf("[Control][%s][%s]", module, mode))
	}
	a = append(a, v...)

	for _, s := range l.sinks {
		s.LogLn(lvl, a...)
	}
}

func (l *Logger) LogLnCurry(lvl int, module string, mode string) func(n int32, v ...interface{}) {
	return func(n int32, v ...interface{}) {
		l.LogLn(lvl, n, module, mode, v...)
	}
}

var LOG Logger

var (
	AppError  func(n int32, v ...interface{}) = LOG.LogLnCurry(1000, "App", "ERR")
	RaftError func(n int32, v ...interface{}) = LOG.LogLnCurry(1000, "Raft", "ERR")
	NetError  func(n int32, v ...interface{}) = LOG.LogLnCurry(1000, "Net", "ERR")

	AppWarn func(n int32, v ...interface{}) = LOG.LogLnCurry(60, "App", "WARN")
	AppInfo func(n int32, v ...interface{}) = LOG.LogLnCurry(50, "App", "INFO")

	RaftWarn func(n int32, v ...interface{}) = LOG.LogLnCurry(40, "Raft", "WARN")
	RaftInfo func(n int32, v ...interface{}) = LOG.LogLnCurry(30, "Raft", "INFO")

	NetWarn func(n int32, v ...interface{}) = LOG.LogLnCurry(20, "Net", "WARN")
	NetInfo func(n int32, v ...interface{}) = LOG.LogLnCurry(10, "Net", "INFO")
)
