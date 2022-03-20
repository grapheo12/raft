package test

import (
	"fmt"
	"raft/pkg/network"
	"raft/pkg/raft"
	"raft/pkg/server"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
)

type cluster struct {
	*testing.T
	// Id of the cluster
	id uint64
	// Port at which the server listens
	port        int
	num_nodes   int
	checkLeak   func()
	testTimeout *time.Timer
	nets        map[int32]*network.Network
	rNodes      map[int32]*raft.RaftNode
	ports       map[int32]string
	servers     map[int32]*server.Server
}

// // Maps a test to a the number of clusters for that test
// var clusters map[string]int
// var clusterCount int
// var net *network.Network

// func init() {
// 	clusters = make(map[string]int)
// 	clusterCount = 0
// 	net := &network.Network{}
// }

var clusterCount int

// create a list of ports
func (c *cluster) createPorts() {
	ports := make([]string, c.num_nodes)
	var port int = 3030

	for i := range ports {
		ports[i] = ":" + fmt.Sprint(port)
		port += 10
	}

	c.ports = make(map[int32]string)
	for i, p := range ports {
		c.ports[int32(i)] = p
	}
}

func (c *cluster) createNets() {
	c.nets = make(map[int32]*network.Network, c.num_nodes)
	for i, p := range c.ports {
		c.nets[i] = &network.Network{}
		err := c.nets[i].Init(p, i)
		if err != nil {
			c.T.Error(err.Error())
		}
	}
	var node_ip string = "127.0.0.1"

	// Create the connections between the networks
	for i, _ := range c.ports {
		for j, p := range c.ports {
			if i != j {
				var addr string = node_ip + p
				_, err := c.nets[i].Connect(j, addr)

				if err != nil {
					c.T.Error(err.Error())
				}
			}
		}
	}
}

func (c *cluster) createRNodes() {
	c.rNodes = make(map[int32]*raft.RaftNode, c.num_nodes)

	for i := range c.ports {
		c.rNodes[i] = &raft.RaftNode{}
		c.rNodes[i].Init(c.nets[i], 100, 200, 300, 400,
			150*time.Millisecond, 300*time.Millisecond, 150*time.Millisecond)
	}

}

func (c *cluster) createServers() {

}

// Create a cluster with an underlying network
func createCluster(t *testing.T) *cluster {
	cPort := 2020 + clusterCount
	clusterCount++

	checkLeak := leaktest.Check(t)
	testTimeout := time.AfterFunc(time.Minute, func() {
		fmt.Printf("test %s timed out, failing...", t.Name())
	})
	heartBeatTimeout := 300 * time.Millisecond

	c := &cluster{
		T:                t,
		id:               uint64(clusterCount),
		port:             cPort,
		checkLeak:        checkLeak,
		testTimeout:      testTimeout,
		heartBeatTimeout: heartBeatTimeout,
	}

	c.createPorts()
	c.createNets()
	c.createRNodes()
	c.createServers()
	return c
}

// Launch a certain number of raft nodes
func (c *cluster) launch(num_nodes int) {

}

func (c *cluster) shutdown() {

}

func TestRaftSingleNode(t *testing.T) {
	c := createCluster(t)
	c.launch(1)
	defer c.shutdown()
}

func TestInit(t *testing.T) {
	defer leaktest.Check(t)()

	go func() {
		for {
			time.Sleep(time.Second)
		}
	}()
}
