package main

// TODO: Move this to a test
// import (
// 	"fmt"
// 	"os"
// 	"raft/internal/lo"
// 	"raft/pkg/network"
// 	"sync"
// 	"time"
// )

// func worker(n *network.Network, n_peers int) {
// 	lo.LOG.AddSink(os.Stdout, 500)
// 	file, _ := os.OpenFile("test.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
// 	lo.LOG.AddSink(file, 0)

// 	ch := make(network.Queue)
// 	n.RegisterQueue(1, ch)
// 	lo.AppInfo(n.NodeId, "Starting operation in 1s")
// 	time.Sleep(time.Second)

// 	snd := 0
// 	rcv := 0

// 	for snd < 100 {
// 		select {
// 		case data := <-ch:
// 			rcv++
// 			lo.AppInfo(n.NodeId, "<- Node", data.NodeId, ": ", string(data.Data))
// 		default:
// 			chk := time.Now()
// 			n.Broadcast(1, []byte(fmt.Sprintf("Greetings from %d", n.NodeId)))
// 			t := time.Since(chk)

// 			lo.AppInfo(n.NodeId, "Msg Time:", t)

// 			snd++
// 		}
// 	}

// 	lo.AppWarn(n.NodeId, "Stopping operation in 2s")
// 	time.Sleep(2 * time.Second)

// }

// func main() {
// 	ports := []string{":2022", ":3022", ":4022"}
// 	nets := make([]*network.Network, len(ports))

// 	for i, p := range ports {
// 		nets[i] = &network.Network{}
// 		err := nets[i].Init(p, int32(i))
// 		if err != nil {
// 			lo.AppError(int32(-1), err.Error())
// 		}
// 	}

// 	time.Sleep(200 * time.Microsecond)

// 	for i := range ports {
// 		for j, p2 := range ports {
// 			if i != j {
// 				_, err := nets[i].Connect(int32(j), "127.0.0.1"+p2)
// 				if err != nil {
// 					lo.AppError(int32(-1), err.Error())
// 				}
// 			}
// 		}
// 	}
// 	time.Sleep(200 * time.Microsecond)

// 	var wg sync.WaitGroup

// 	for i := range ports {
// 		wg.Add(1)
// 		go func(id int) {
// 			worker(nets[id], len(ports))
// 			wg.Done()
// 		}(i)
// 	}

// 	wg.Wait()

// 	for i := range ports {
// 		nets[i].StopServer()
// 	}

// }
