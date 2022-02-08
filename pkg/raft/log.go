package raft

// 4 types of msg : VoteRequest, VoteResponse, LogRequest, LogResponse

type logEntry struct {
}

type log struct {
	entry logEntry
	next  *logEntry
}
