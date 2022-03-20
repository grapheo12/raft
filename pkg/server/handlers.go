package server

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"raft/pkg/raft"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type ReadResp struct {
	Msg []string `json:"message"`
}

type WriteReq struct {
	Content string `json:"content"`
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
		Msg: s.DB.GetAll(),
	}

	data, _ := json.Marshal(resp)
	w.Write(data)
}

func (s *Server) raftWrite(w http.ResponseWriter, r *http.Request) {
	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-type", "application/json")
		w.Write([]byte("{\"message\": \"No request body\"}"))
		return
	}
	req := WriteReq{}
	err = json.Unmarshal(raw, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-type", "application/json")
		w.Write([]byte("{\"message\": \"Bad request body\"}"))
		return
	}

	if s.RNode.State != raft.LEADER {
		leader := s.Peers[s.RNode.CurrLeaderId]
		w.WriteHeader(http.StatusBadGateway)
		w.Header().Set("Content-type", "application/json")
		w.Write([]byte(`{
			"message": "Not Leader",
			"leader": "` + leader + `"
		}`))
		return
	}

	timeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	var wg sync.WaitGroup
	written := false

	defer cancel()
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-timeout.Done():
			return
		case s.RNode.ClientIn <- []byte(req.Content):
			written = true
			return
		}
	}()

	wg.Wait()

	if !written {
		w.WriteHeader(http.StatusRequestTimeout)
		w.Header().Set("Content-type", "application/json")
		w.Write([]byte("{\"message\": \"Timeout while writing\"}"))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-type", "application/json")
	w.Write([]byte("{\"message\": \"Write successful\"}"))
}

func (s *Server) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/ping", s.ping).Methods("GET")
	router.HandleFunc("/read", s.raftRead).Methods("GET")
	router.HandleFunc("/write", s.raftWrite).Methods("POST")
}
