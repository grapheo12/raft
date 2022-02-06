package main

import (
	"fmt"
	"raft/pkg/network"
	"sync"
	"time"
)

func worker(n *network.Network, n_peers int) {
	ch := make(network.Queue)
	n.RegisterQueue(1, ch)
	fmt.Println("Node", n.NodeId, "| ", "Starting operation in 1s")
	time.Sleep(time.Second)

	snd := 0
	rcv := 0

	for snd < 100 {
		select {
		case data := <-ch:
			rcv++
			fmt.Println("Node", n.NodeId, "-> Node", data.NodeId, ": ", string(data.Data))
		default:
			chk := time.Now()
			n.Broadcast(1, []byte(fmt.Sprintf("Greetings from %d", n.NodeId)))
			t := time.Since(chk)

			fmt.Println("Node", n.NodeId, "| Msg Time:", t)

			snd++
		}
	}

	fmt.Println("Node", n.NodeId, "| ", "Stopping operation in 2s")
	time.Sleep(2 * time.Second)

}

func main() {
	ports := []string{":2022", ":3022", ":4022"}
	nets := make([]*network.Network, len(ports))

	for i, p := range ports {
		nets[i] = &network.Network{}
		err := nets[i].Init(p, int32(i))
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	time.Sleep(200 * time.Microsecond)

	for i := range ports {
		for j, p2 := range ports {
			if i != j {
				_, err := nets[i].Connect(int32(j), "127.0.0.1"+p2)
				if err != nil {
					fmt.Println(err.Error())
				}
			}
		}
	}
	time.Sleep(200 * time.Microsecond)

	var wg sync.WaitGroup

	for i := range ports {
		wg.Add(1)
		go func(id int) {
			worker(nets[id], len(ports))
			wg.Done()
		}(i)
	}

	wg.Wait()

	for i := range ports {
		nets[i].StopServer()
	}

}
