package network

import (
	"errors"
	"fmt"
	"net"
	"raft/pkg/rpc"
)

func (n *Network) Connect(nodeId int32, addr string) (net.Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	me := rpc.InitMessage{
		Magic:  0xdead,
		NodeId: n.NodeId,
	}

	b, _ := me.Marshal()

	fmt.Println("Node", n.NodeId, "| Sending Length:", len(b))
	conn.Write(b)

	n.Peers[nodeId] = addr
	n.OutConn[nodeId] = conn

	return conn, nil
}

func (n *Network) Send(nodeId int32, qId int32, data []byte) error {
	conn, ok := n.OutConn[nodeId]
	if !ok {
		return errors.New("unregistered peer")
	}

	msg := rpc.NodeMessage{
		Data:   data,
		QId:    qId,
		NodeId: n.NodeId,
	}

	b, err := msg.Marshal()
	if err != nil {
		return err
	}

	// lenData := rpc.MessageLen{
	// 	Magic:  0xcaffe,
	// 	Length: int32(len(b)),
	// }
	// b2, err := lenData.Marshal()
	// if err != nil {
	// 	return err
	// }

	// _, err = conn.Write(b2)
	// if err != nil {
	// 	return err
	// }

	_, err = conn.Write(b)
	if err != nil {
		return err
	}

	return nil

}

func (n *Network) Broadcast(qId int32, data []byte) {
	for nId := range n.OutConn {
		n.Send(nId, qId, data)
	}
}
