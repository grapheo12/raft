package test

import (
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
)

// // Maps a test to a the number of clusters for that test
// var clusters map[string]int
// var clusterCount int
// var net *network.Network

// func init() {
// 	clusters = make(map[string]int)
// 	clusterCount = 0
// 	net := &network.Network{}
// }

func TestRaftSingleNode(t *testing.T) {
	defer leaktest.Check(t)()
	c := createCluster(t, 1)
	c.shutdown()
}

// Check that leader is elected and only a single leader exists
func TestElectionSimple(t *testing.T) {
	c := createCluster(t, 3)
	defer c.shutdown()
	c.checkSingleLeader()
}

func TestElectionWithDisconnect(t *testing.T) {
	// lo.LOG.AddSink(os.Stdout, 25)
	c := createCluster(t, 3)
	defer c.shutdown()

	leaderId, leaderTerm := c.checkSingleLeader()
	c.T.Logf("Current Leader %d", leaderId)
	c.RemoveNode(leaderId)

	time.Sleep(400 * time.Millisecond)

	newLeaderId, newLeaderTerm := c.checkSingleLeader()
	c.T.Logf("New Leader %d", newLeaderId)
	if newLeaderId == leaderId {
		c.T.Errorf("Leader needs to be different post disconnect! Election not working")
	}

	if newLeaderTerm <= leaderTerm {
		c.T.Errorf("Term of new Leader should be >= orginal term. Instead got %d and %d", newLeaderTerm, leaderTerm)
	}
}

func TestInit(t *testing.T) {
	defer leaktest.Check(t)()

	go func() {
		for {
			time.Sleep(time.Second)
		}
	}()
}
