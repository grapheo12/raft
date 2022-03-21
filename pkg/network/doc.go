// network package provides the abstraction of an overlay P2P network.
//
// It listens on a port for incoming messages
// and establishes persistent TCP connections with its peers.
// Incoming messages have designated channels (qId) where they are redirected.
// Other layers depending on the network layer are supposed to implement their own buffering scheme
// which will listen on the other end of these channels.
//
// We use persistent TCP connections to avoid the time required to create a new connection for every message.
// This also allows us to take advantage of congestion control and message retrying facilities of TCP.
// Stale connections are re-established if detected.
package network
