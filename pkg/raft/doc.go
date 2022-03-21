// raft package provides the implementation of the raft protocol.
//
// We have implemented only the basic protocol
// with experimental verification of safety and liveness.
// The implementation is based on State Machine Representation of Raft.
// A central `NodeMain` function runs in loop and in each iteration
// performs a step as per its state (Leader, Follower or Candidate).
package raft
