package tests

import (
	"os"
	"raft/internal/lo"
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

// Test that on leader's disconnection a new leader is eventually elected
func TestElectionWithLeaderDisconnect(t *testing.T) {
	lo.LOG.AddSink(os.Stdout, 25)
	c := createCluster(t, 3)
	defer leaktest.Check(t)()
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

// Test that leader disconnect , followed by another disconnect leads to no quorum being formed
func TestElectionWithLeaderAndOtherDisconnect(t *testing.T) {
	c := createCluster(t, 3)
	defer leaktest.Check(t)()
	defer c.shutdown()

	leaderId, _ := c.checkSingleLeader()
	c.RemoveNode(leaderId)
	other_node_id := (leaderId + 1) % 3
	c.RemoveNode(other_node_id)

	// NO quorum can be attained
	time.Sleep(400 * time.Millisecond)
	c.checkNoLeader()
}
