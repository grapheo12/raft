package config

import "time"

type ConfigSet struct {
	NetworkMembers []string
	Id             int32
	NetworkIp      string
	NetworkPort    string

	EMinT time.Duration
	EMaxT time.Duration

	ClientIp      string
	ClientPort    string
	ClientMembers []string

	LogLevel int
}

func ParseConfigs(src []string) ConfigSet {
	cs, inisrc := ParseArgs(src)
	ParseIni(inisrc, &cs)

	return cs
}
