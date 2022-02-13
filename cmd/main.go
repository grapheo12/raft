package main

import (
	"context"
	"os"
	"raft/internal/lo"
	"raft/pkg/network"
	"raft/pkg/raft"
	"raft/pkg/server"
	"time"
)

func main() {
	lo.LOG.AddSink(os.Stdout, 25)

	ports := []string{":2022", ":3022", ":4022"}
	nets := make([]*network.Network, len(ports))

	for i, p := range ports {
		nets[i] = &network.Network{}
		err := nets[i].Init(p, int32(i))
		if err != nil {
			lo.AppError(int32(-1), err.Error())
		}
	}

	time.Sleep(200 * time.Microsecond)

	for i := range ports {
		for j, p2 := range ports {
			if i != j {
				_, err := nets[i].Connect(int32(j), "127.0.0.1"+p2)
				if err != nil {
					lo.AppError(int32(-1), err.Error())
				}
			}
		}
	}
	time.Sleep(200 * time.Microsecond)

	rNodes := make([]*raft.RaftNode, len(ports))
	for i := range ports {
		rNodes[i] = &raft.RaftNode{}
		rNodes[i].Init(nets[i], 100, 200, 300, 400,
			150*time.Millisecond, 300*time.Millisecond, 150*time.Millisecond)
	}
	time.Sleep(200 * time.Microsecond)

	serverPorts := []string{":2020", ":3020", ":4020"}
	servers := make([]*server.Server, len(serverPorts))
	peers := make(map[int32]string)
	for i := range serverPorts {
		peers[int32(i)] = "127.0.0.1" + serverPorts[i]
	}

	for i := range serverPorts {
		servers[i] = &server.Server{}
		servers[i].Init(serverPorts[i], rNodes[i], int32(i), peers)
	}

	time.Sleep(200 * time.Microsecond)

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	go func() {
		<-sigs
		for i := range ports {
			nets[i].StopServer()
			servers[i].Srv.Shutdown(context.Background())
		}
		done <- true
	}()
	<-done
	lo.AppInfo(int32(-1), "Exiting")
}
