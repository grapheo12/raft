package tests

import (
	"fmt"
	"raft/internal/lo"
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

// Creates the raft nodes
func (c *cluster) createRNodes() {
	c.rNodes = make(map[int32]*raft.RaftNode, c.num_nodes)

	for i := range c.ports {
		c.rNodes[i] = &raft.RaftNode{}
		c.rNodes[i].Init(c.nets[i], 100, 200, 300, 400,
			150*time.Millisecond, 300*time.Millisecond, 150*time.Millisecond)
	}

}

// Creates the servers for the Raft Nodes
func (c *cluster) createServers() {
	serverPorts := make([]string, c.num_nodes)

	var port int = 3032

	for i := range serverPorts {
		serverPorts[i] = ":" + fmt.Sprint(port)
		port += 10
	}

	c.servers = make(map[int32]*server.Server, c.num_nodes)
	peers := make(map[int32]string)

	for i, p := range serverPorts {
		peers[int32(i)] = "127.0.0.1" + p
	}

	for i := range serverPorts {
		c.servers[int32(i)] = &server.Server{}
		c.servers[int32(i)].Init(serverPorts[i], c.rNodes[int32(i)], int32(i), peers)
	}
}

// Create a cluster containing a given number of raft nodes
func createCluster(t *testing.T, num_nodes int) *cluster {
	cPort := 2020 + clusterCount
	clusterCount++

	checkLeak := leaktest.Check(t)
	testTimeout := time.AfterFunc(time.Minute, func() {
		fmt.Printf("test %s timed out, failing...", t.Name())
	})

	c := &cluster{
		T:           t,
		id:          uint64(clusterCount),
		port:        cPort,
		checkLeak:   checkLeak,
		testTimeout: testTimeout,
		num_nodes:   num_nodes,
	}

	c.createPorts()
	c.createNets()
	c.createRNodes()
	c.createServers()
	return c
}

// Shoutsdown a particular server
func (c *cluster) shutdown() {
	for i := range c.ports {
		c.nets[i].StopServer()
		c.servers[i].Shutdown()
	}
}

// Checks that whether the cluster has a Leader or not
func (c *cluster) checkSingleLeader() (int, int) {
	for r := 0; r < 5; r++ {
		// Check for five rounds
		leaderId := -1
		leaderTerm := -1
		for i, rNode := range c.rNodes {
			if rNode.State == raft.LEADER {
				if leaderId < 0 {
					leaderId = int(i)
					leaderTerm = int(rNode.Term)
				} else {
					c.T.Errorf("Two leaders exist!! Both %d and %d  are leaders", leaderId, i)
				}
			}
		}
		if leaderId != -1 {
			return leaderId, leaderTerm
		}

		time.Sleep(200 * time.Millisecond)
	}
	c.T.Errorf("No leader found for five consecutive checks!!!")
	return -1, -1
}

func (c *cluster) RemoveNode(node_id int) {
	c.T.Logf("Node_id received is %d", node_id)

	lo.RaftInfo(int32(node_id), "Stopped the node")
	c.rNodes[int32(node_id)].StopNode()
	c.nets[int32(node_id)].StopServer()
	c.servers[int32(node_id)].Shutdown()
	delete(c.rNodes, int32(node_id))
	delete(c.nets, int32(node_id))
	delete(c.servers, int32(node_id))
	delete(c.ports, int32(node_id))
	c.num_nodes--
}
