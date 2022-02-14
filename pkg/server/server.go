package server

import (
	"context"
	"net/http"
	"raft/internal/lo"
	"raft/pkg/raft"

	"github.com/gorilla/mux"
)

type Server struct {
	Port  string
	nId   int32
	RNode *raft.RaftNode
	Peers map[int32]string // nodeId -> server's IP:addr
	Srv   *http.Server

	DB      *Buffer
	dbClose context.CancelFunc
}

func (s *Server) Init(port string, rNode *raft.RaftNode, nId int32, peers map[int32]string) {
	s.Peers = peers
	s.Port = port
	s.RNode = rNode
	s.nId = nId

	s.DB = &Buffer{}
	s.DB.Init(rNode)
	ctx, cancel := context.WithCancel(context.Background())
	s.dbClose = cancel
	go s.DB.FetchDataWorker(ctx)

	router := mux.NewRouter()
	s.RegisterRoutes(router)

	srv := http.Server{
		Addr:    port,
		Handler: router,
	}

	s.Srv = &srv

	go func(s *http.Server) {
		err := s.ListenAndServe()

		if err != nil {
			lo.AppError(nId, "Client stopped")
		}
	}(&srv)

}

func (s *Server) Shutdown() {
	s.Srv.Shutdown(context.Background())
	s.dbClose()
}
