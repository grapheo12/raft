package server

import (
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
}

func (s *Server) Init(port string, rNode *raft.RaftNode, nId int32, peers map[int32]string) {
	s.Peers = peers
	s.Port = port
	s.RNode = rNode
	s.nId = nId

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
