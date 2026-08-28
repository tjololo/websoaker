package cmd

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/tjololo/websoaker/internal/server"
)

type SinkServer struct {
	mux      sync.Mutex
	reqCount float64
}

func (s *SinkServer) StartSinkServer(port string) {
	s.reqCount = 0
	slog.Info("Starting sink server")
	http.HandleFunc("/ping", s.pingHandler)
	http.HandleFunc("/status", s.statusHandler)
	server.ServeGraceful(port)
}

func (s *SinkServer) pingHandler(w http.ResponseWriter, _ *http.Request) {
	slog.Debug("Ping received")
	s.IncCounter()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("{\"status\": \"ok\"}"))
	if err != nil {
		slog.Warn("Failed to write response", "error", err)
	}
}

func (s *SinkServer) statusHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(fmt.Sprintf("{\"reqCount\": \"%.0f\"}", s.reqCount)))
	if err != nil {
		slog.Warn("Failed to write response", "error", err)
	}
}

func (s *SinkServer) IncCounter() {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.reqCount++
}
