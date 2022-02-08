package network

import (
	"context"
	"errors"
	"net"
	"raft/internal/lo"
	"raft/pkg/rpc"
	"reflect"
)

func (n *Network) acceptor(lr net.Listener, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			lr.Close()
			return
		default:
			lo.NetInfo(n.NodeId, "Ready to Accept")
			conn, err := lr.Accept()
			if err != nil {
				continue
			}
			b := make([]byte, 20)
			b_len, err := conn.Read(b)

			if err != nil {
				lo.NetWarn(n.NodeId, err.Error())
				conn.Close()
				continue
			}
			initData := rpc.InitMessage{}
			err = initData.Unmarshal(b[:b_len])
			if err != nil {
				lo.NetWarn(n.NodeId, err.Error(), "Len:", b_len, "Data:", b[:b_len])
				conn.Close()
				continue
			}
			lo.NetInfo(n.NodeId, "ConnData:", initData)
			n.newConn <- ConnInit{
				NodeId: initData.NodeId,
				Conn:   conn,
			}
			lo.NetInfo(n.NodeId, "Connected to", initData.NodeId)
		}
	}
}

func (n *Network) receiver(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			n.inConnLck.RLock()
			cases := make([]reflect.SelectCase, len(n.InConn)+1)
			i := 0
			for _, v := range n.InConn {
				cases[i] = reflect.SelectCase{
					Dir:  reflect.SelectRecv,
					Chan: reflect.ValueOf(v),
				}
				i++
			}
			cases[i] = reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(n.newConnSignal),
			}
			n.inConnLck.RUnlock()

			chosen, val, recvOk := reflect.Select(cases)
			if !recvOk {
				continue
			}

			if chosen == i {
				continue
			}

			msg := val.Interface().(rpc.NodeMessage)
			lo.NetInfo(n.NodeId, "Receiving message from", msg.NodeId, "for QId", msg.QId)
			ch, ok := n.Queues[msg.QId]
			if !ok {
				continue
			}
			ch <- msg
		}
	}
}

func (n *Network) reader(conn net.Conn, sender chan rpc.NodeMessage, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			dataBytes := make([]byte, MAX_MSGLEN)
			b_len, err := conn.Read(dataBytes)
			if err != nil {
				lo.NetWarn(n.NodeId, err.Error())
				continue
			}

			msg := rpc.NodeMessage{}
			err = msg.Unmarshal(dataBytes[:b_len])
			if err != nil {
				lo.NetWarn(n.NodeId, err.Error())
				continue
			}

			sender <- msg
		}
	}
}

func (n *Network) Server(lr net.Listener, ctx context.Context) {
	ctxAcc, endAcc := context.WithCancel(ctx)
	go n.acceptor(lr, ctxAcc)

	ctxRecv, endRecv := context.WithCancel(ctx)
	go n.receiver(ctxRecv)

	for {
		select {
		case <-ctx.Done():
			endAcc()
			endRecv()
			lo.NetWarn(n.NodeId, "Server exiting")
			return
		case connData := <-n.newConn:
			lo.NetInfo(n.NodeId, "Adding Reader for new Connection:", connData)
			ch := make(chan rpc.NodeMessage)
			ktx, _ := context.WithCancel(ctx)
			go n.reader(connData.Conn, ch, ktx)
			n.inConnLck.Lock()
			n.InConn[connData.NodeId] = ch
			n.inConnLck.Unlock()
			n.newConnSignal <- true
		case <-n.msgSignal:
			continue
		}
	}
}

func (n *Network) RegisterQueue(qId int32, ch Queue) error {
	_, ok := n.Queues[qId]
	if ok {
		return errors.New("qId already registered")
	}

	n.Queues[qId] = ch

	return nil
}
