package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", "application/json")

	w.Write([]byte("{\"message\": \"pong\"}"))
}

func (s *Server) raftRead(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", "application/json")

	w.Write([]byte("{\"message\": \"pong\"}"))
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
