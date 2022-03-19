package config

import (
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

func ParseIni(src string, cs *ConfigSet) {
	cfg, err := ini.Load(src)
	if err != nil {
		panic(err)
	}

	if cs.NetworkMembers == nil {
		members := cfg.Section("network").Key("network-members").MustString("")
		if members == "" {
			panic("Network members not specified")
		}
		arr := strings.Split(members, ",")
		cs.NetworkMembers = make([]string, 0)
		for _, a := range arr {
			cs.NetworkMembers = append(cs.NetworkMembers, strings.Trim(a, " "))
		}
	}

	if cs.ClientMembers == nil {
		members := cfg.Section("client").Key("client-members").MustString("")
		if members == "" {
			panic("Client members not specified")
		}
		arr := strings.Split(members, ",")
		cs.ClientMembers = make([]string, 0)
		for _, a := range arr {
			cs.ClientMembers = append(cs.ClientMembers, strings.Trim(a, " "))
		}
	}

	if cs.Id == -1 {
		cs.Id = int32(cfg.Section("network").Key("id").MustInt(-1))
		if cs.Id == -1 {
			panic("Id not specified")
		}
	}
	if cs.NetworkIp == "" {
		cs.NetworkIp = cfg.Section("network").Key("network-ip").MustString("")
		if cs.NetworkIp == "" {
			panic("Network-ip not specified")
		}
	}
	if cs.NetworkPort == "" {
		cs.NetworkIp = cfg.Section("network").Key("network-port").MustString("")
		if cs.NetworkPort == "" {
			panic("Network-port not specified")
		}
	}

	if cs.EMinT == -1*time.Millisecond {
		cs.EMinT = time.Duration(cfg.Section("raft").Key("min-election-timeout").MustInt(-1) * int(time.Millisecond))
		if cs.EMinT == -1*time.Millisecond {
			panic("min-election-timeout not specified")
		}
	}
	if cs.EMaxT == -1*time.Millisecond {
		cs.EMaxT = time.Duration(cfg.Section("raft").Key("max-election-timeout").MustInt(-1) * int(time.Millisecond))
		if cs.EMaxT == -1*time.Millisecond {
			panic("max-election-timeout not specified")
		}
	}

	if cs.ClientIp == "" {
		cs.ClientIp = cfg.Section("client").Key("client-ip").MustString("")
		if cs.ClientIp == "" {
			panic("Client-ip not specified")
		}
	}
	if cs.ClientPort == "" {
		cs.ClientPort = cfg.Section("client").Key("client-port").MustString("")
		if cs.ClientPort == "" {
			panic("Client-port not specified")
		}
	}

	if cs.LogLevel == -1 {
		cs.LogLevel = cfg.Section("log").Key("log-level").MustInt(25)
	}

}
