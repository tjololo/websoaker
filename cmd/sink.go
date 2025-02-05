package cmd

import (
	"fmt"
	"github.com/tjololo/websoaker/internal/server"
	"log"
	"net/http"
	"sync"
)

type SinkServer struct {
	mux      sync.Mutex
	reqCount float64
}

func (s *SinkServer) StartSinkServer(port string) {
	s.reqCount = 0
	log.Println("Starting sink server")
	http.HandleFunc("/ping", s.pingHandler)
	http.HandleFunc("/status", s.statusHandler)
	server.ServeGraceful(port)
}

func (s *SinkServer) pingHandler(w http.ResponseWriter, _ *http.Request) {
	log.Println("Ping received")
	s.IncCounter()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("{\"status\": \"ok\"}"))
	if err != nil {
		log.Printf("Failed to write response %v", err)
	}
}

func (s *SinkServer) statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(fmt.Sprintf("{\"reqCount\": \"%.0f\"}", s.reqCount)))
	if err != nil {
		log.Printf("Failed to write response %v", err)
	}
}

func (s *SinkServer) IncCounter() {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.reqCount++
}
