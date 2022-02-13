package server

import (
	"encoding/json"
	"net/http"
	"raft/pkg/raft"

	"github.com/gorilla/mux"
)

type ReadResp struct {
	Msg raft.LogArrayType `json:"message"`
}

func (s *Server) ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", "application/json")

	w.Write([]byte("{\"message\": \"pong\"}"))
}

func (s *Server) raftRead(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", "application/json")

	resp := ReadResp{
		Msg: s.RNode.Log.LogArray,
	}

	data, _ := json.Marshal(resp)
	w.Write(data)
}

func (s *Server) raftWrite(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", "application/json")

	w.Write([]byte("{\"message\": \"pong\"}"))
}

func (s *Server) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/ping", s.ping).Methods("GET")
	router.HandleFunc("/read", s.raftRead).Methods("GET")
	router.HandleFunc("/write", s.raftWrite).Methods("POST")
}
