package network

import (
	"context"
	"net"
	"raft/pkg/rpc"
	"sync"
)

const MAX_MSGLEN = 1000

type Queue chan rpc.NodeMessage

type ConnInit struct {
	NodeId int32
	Conn   net.Conn
}

type Network struct {
	Port   string
	NodeId int32

	Peers   map[int32]string               // nodeId -> addr:port
	OutConn map[int32]net.Conn             // nodeId -> open TCP sockets
	InConn  map[int32]chan rpc.NodeMessage // nodeId -> data sent
	Queues  map[int32]Queue                // qId -> Queue

	StopServer    context.CancelFunc
	newConn       chan ConnInit
	newConnSignal chan bool
	msgSignal     chan bool

	inConnLck sync.RWMutex
}

func (n *Network) Init(port string, nodeId int32) error {
	n.Port = port
	n.NodeId = nodeId

	n.Peers = make(map[int32]string)
	n.OutConn = make(map[int32]net.Conn)
	n.InConn = make(map[int32]chan rpc.NodeMessage)
	n.Queues = make(map[int32]Queue)

	listener, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	go n.Server(listener, ctx)

	n.StopServer = cancel

	n.msgSignal = make(chan bool)
	n.newConn = make(chan ConnInit)
	n.newConnSignal = make(chan bool)

	return nil

}
