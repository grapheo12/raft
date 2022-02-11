package raft

type VoteRequestMsg struct {
	CandidateId      int32
	CandidateTerm    int
	CandidateLogLen  int
	CandidateLogTerm int
}

type VoteResponseMsg struct {
	VoterId   int32
	VoterTerm int
	Granted   bool
}

type LogRequestMsg struct {
	LeaderId   int32
	LeaderTerm int
	PrefixLen  int
	PrefixTerm int
	CommitLen  int
	Suffix     *LogEntry
}

type LogResponseMsg struct {
	FollowerId       int32
	FollowerTerm     int
	LogCommitAck     int
	LogCommitSuccess bool
}
