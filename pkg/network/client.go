package network

import (
	"errors"
	"net"
	"raft/internal/lo"
	"raft/pkg/rpc"
	"sync"
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

	lo.NetInfo(n.NodeId, "Sending Length:", len(b))
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
	var wg sync.WaitGroup

	for nId := range n.OutConn {
		wg.Add(1)

		go func(_n, q int32, d []byte) {
			n.Send(_n, q, d)
			wg.Done()
		}(nId, qId, data)
	}

	wg.Wait()
}
