package main

import (
	"os"
	"raft/internal/config"
	"raft/internal/lo"
	"raft/pkg/network"
	"raft/pkg/raft"
	"raft/pkg/server"
	"time"
)

func main() {
	cfg := config.ParseConfigs(os.Args)

	net := &network.Network{}
	err := net.Init(cfg.NetworkPort, cfg.Id)
	if err != nil {
		lo.AppError(int32(-1), err.Error())
	}

	time.Sleep(200 * time.Microsecond)

	for idx, member := range cfg.NetworkMembers {
		if idx != int(cfg.Id) {
			_, err := net.Connect(int32(idx), member)
			if err != nil {
				lo.AppError(int32(-1), err.Error())
			}
		}
	}

	time.Sleep(200 * time.Microsecond)

	rNode := &raft.RaftNode{}
	rNode.Init(net, 100, 200, 300, 400, cfg.EMinT, cfg.EMaxT, cfg.EMinT)

	time.Sleep(200 * time.Microsecond)

	peers := make(map[int32]string)
	for i := range cfg.ClientMembers {
		peers[int32(i)] = cfg.ClientMembers[i]
	}

	server := &server.Server{}
	server.Init(cfg.ClientPort, rNode, cfg.Id, peers)

	time.Sleep(200 * time.Microsecond)

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	go func() {
		<-sigs
		net.StopServer()
		server.Shutdown()
		done <- true
	}()
	<-done
	lo.AppInfo(int32(-1), "Exiting")

}
