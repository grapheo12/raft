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
	})
	network_ip := parser.String("", "network-ip", &argparse.Options{
		Required: false,
		Help:     "Ip to listen to for Raft messages, may be different from above",
	})
	network_port := parser.String("", "network-port", &argparse.Options{
		Required: false,
		Help:     "Port to listen to for Raft messages, may be different from above",
	})

	min_election_timeout := parser.Int("", "min-election-timeout", &argparse.Options{
		Required: false,
		Help:     "time in ms",
	})
	max_election_timeout := parser.Int("", "max-election-timeout", &argparse.Options{
		Required: false,
		Help:     "time in ms",
	})

	client_ip := parser.String("", "client-ip", &argparse.Options{
		Required: false,
		Help:     "Ip to listen to for client messages",
	})
	client_port := parser.String("", "client-port", &argparse.Options{
		Required: false,
		Help:     "Port to listen to for client messages",
	})
	client_members := parser.String("", "client-members", &argparse.Options{
		Required: false,
		Help:     "comma separated list of ip:port for client messages, containing self",
	})
	log_level := parser.Int("l", "log-level", &argparse.Options{
		Required: false,
		Help:     "Loglevel to use",
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

		LogLevel: -1,
	}

	if network_members != nil {
		arr := strings.Split(*network_members, ",")
		cs.NetworkMembers = make([]string, 0)
		for _, a := range arr {
			cs.NetworkMembers = append(cs.NetworkMembers, strings.Trim(a, " "))
		}
	}

	if client_members != nil {
		arr := strings.Split(*client_members, ",")
		cs.ClientMembers = make([]string, 0)
		for _, a := range arr {
			cs.ClientMembers = append(cs.ClientMembers, strings.Trim(a, " "))
		}
	}

	if id != nil {
		cs.Id = int32(*id)
	}

	if network_ip != nil {
		cs.NetworkIp = *network_ip
	}

	if network_port != nil {
		cs.NetworkPort = *network_port
	}

	if min_election_timeout != nil {
		cs.EMinT = time.Duration((*min_election_timeout) * int(time.Millisecond))
	}

	if max_election_timeout != nil {
		cs.EMaxT = time.Duration((*max_election_timeout) * int(time.Millisecond))
	}

	if client_ip != nil {
		cs.ClientIp = *client_ip
	}

	if client_port != nil {
		cs.ClientPort = *client_port
	}

	if log_level != nil {
		cs.LogLevel = *log_level
	}

	return cs, *inisrc

}
