package config

import (
	"strings"
	"time"

	"github.com/akamensky/argparse"
)

func ParseArgs(src []string) (ConfigSet, string) {
	parser := argparse.NewParser("", "Raft Server")

	inisrc := parser.String("c", "config", &argparse.Options{
		Required: true,
		Help:     "config.ini source",
	})

	network_members := parser.String("", "network-members", &argparse.Options{
		Required: false,
		Help:     "comma separated list of ip:port for Raft messages, containing self",
	})

	id := parser.Int("", "id", &argparse.Options{
		Required: false,
		Help:     "Index of self in the above list",
		Default:  -1,
	})
	network_ip := parser.String("", "network-ip", &argparse.Options{
		Required: false,
		Help:     "Ip to listen to for Raft messages, may be different from above",
		Default:  "",
	})
	network_port := parser.String("", "network-port", &argparse.Options{
		Required: false,
		Help:     "Port to listen to for Raft messages, may be different from above",
		Default:  "",
	})

	min_election_timeout := parser.Int("", "min-election-timeout", &argparse.Options{
		Required: false,
		Help:     "time in ms",
		Default:  -1,
	})
	max_election_timeout := parser.Int("", "max-election-timeout", &argparse.Options{
		Required: false,
		Help:     "time in ms",
		Default:  -1,
	})

	client_ip := parser.String("", "client-ip", &argparse.Options{
		Required: false,
		Help:     "Ip to listen to for client messages",
		Default:  "",
	})
	client_port := parser.String("", "client-port", &argparse.Options{
		Required: false,
		Help:     "Port to listen to for client messages",
		Default:  "",
	})
	client_members := parser.String("", "client-members", &argparse.Options{
		Required: false,
		Help:     "comma separated list of ip:port for client messages, containing self",
		Default:  "",
	})
	log_level := parser.Int("l", "log-level", &argparse.Options{
		Required: false,
		Help:     "Loglevel to use",
		Default:  -1,
	})
	warmup_time := parser.Int("w", "warmup-time", &argparse.Options{
		Required: false,
		Help:     "Time to wait before starting operations (in ms)",
		Default:  1000,
	})

	err := parser.Parse(src)
	if err != nil {
		panic(parser.Usage(err))
	}

	cs := ConfigSet{
		NetworkMembers: nil,
		Id:             -1,
		NetworkIp:      "",
		NetworkPort:    "",

		EMinT: -1 * time.Millisecond,
		EMaxT: -1 * time.Millisecond,

		ClientIp:      "",
		ClientPort:    "",
		ClientMembers: nil,

		LogLevel:   -1,
		WarmupTime: time.Duration(*warmup_time * int(time.Millisecond)),
	}

	if *network_members != "" {
		arr := strings.Split(*network_members, ",")
		cs.NetworkMembers = make([]string, 0)
		for _, a := range arr {
			cs.NetworkMembers = append(cs.NetworkMembers, strings.Trim(a, " "))
		}
	}

	if *client_members != "" {
		arr := strings.Split(*client_members, ",")
		cs.ClientMembers = make([]string, 0)
		for _, a := range arr {
			cs.ClientMembers = append(cs.ClientMembers, strings.Trim(a, " "))
		}
	}

	if *id != -1 {
		cs.Id = int32(*id)
	}

	if *network_ip != "" {
		cs.NetworkIp = *network_ip
	}

	if *network_port != "" {
		cs.NetworkPort = *network_port
	}

	if *min_election_timeout != -1 {
		cs.EMinT = time.Duration((*min_election_timeout) * int(time.Millisecond))
	}

	if *max_election_timeout != -1 {
		cs.EMaxT = time.Duration((*max_election_timeout) * int(time.Millisecond))
	}

	if *client_ip != "" {
		cs.ClientIp = *client_ip
	}

	if *client_port != "" {
		cs.ClientPort = *client_port
	}

	if *log_level != -1 {
		cs.LogLevel = *log_level
	}

	return cs, *inisrc

}
