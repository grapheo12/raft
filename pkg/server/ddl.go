package server

import (
	"context"
	"raft/pkg/raft"
	"sync"
)

type Buffer struct {
	Msgs  []string
	rNode *raft.RaftNode
	lck   sync.RWMutex
}

func (b *Buffer) Init(r *raft.RaftNode) {
	b.Msgs = make([]string, 0)
	b.rNode = r
}

func (b *Buffer) FetchDataWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-b.rNode.ClientOut:
			b.lck.Lock()
			b.Msgs = append(b.Msgs, string(data))
			b.lck.Unlock()
		}
	}
}

func (b *Buffer) GetAll() []string {
	b.lck.RLock()
	defer b.lck.RUnlock()

	return b.Msgs[:]
}
