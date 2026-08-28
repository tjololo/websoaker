package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/tjololo/websoaker/internal/server"
)

type SourceServer struct {
	running      bool
	concurrency  int
	notifyChan   chan bool
	soakHost     string
	mux          sync.Mutex
	successCount float64
	failedCount  float64
	maxCons      int
	basePath     string
}

func NewSourceServer(soakHost string, basePath string, concurrency int, maxCons int) *SourceServer {
	return &SourceServer{
		running:     false,
		concurrency: concurrency,
		notifyChan:  make(chan bool),
		soakHost:    soakHost,
		maxCons:     maxCons,
		basePath:    basePath,
	}
}

func (s *SourceServer) StartSourceServer(port string) {
	slog.Info("Starting source server", "port", port, "concurrency", s.concurrency)
	http.HandleFunc("/start", s.startHandler)
	http.HandleFunc("/stop", s.stopHandler)
	http.HandleFunc("/status", s.statusHandler)
	server.ServeGraceful(port)
}

func (s *SourceServer) startHandler(w http.ResponseWriter, _ *http.Request) {
	if s.running {
		w.WriteHeader(http.StatusConflict)
		_, err := w.Write([]byte("{\"status\": \"Server already running\"}"))
		if err != nil {
			slog.Warn("Failed to write response", "error", err)
		}
		return
	}
	slog.Info("Start request received", "soakHost", s.soakHost, "concurrency", s.concurrency)
	s.running = true
	transport := &http.Transport{
		MaxConnsPerHost:     s.maxCons,
		MaxIdleConnsPerHost: s.maxCons,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}
	go func() {
		for {
			select {
			case <-s.notifyChan:
				s.running = false
				slog.Info("Stop notification received")
				return
			default:
				guard := make(chan struct{}, s.concurrency)
				wg := &sync.WaitGroup{}
				for i := 0; i < s.concurrency; i++ {
					wg.Add(1)
					guard <- struct{}{}
					go func(n int) {
						resp, err := client.Get(fmt.Sprintf("%s%s/ping", s.soakHost, s.basePath))
						if err != nil {
							slog.Error("Error making request", "error", err)
							s.incFailed()
						} else {
							_, httpErr := io.Copy(io.Discard, resp.Body)
							if httpErr != nil {
								slog.Warn("Error reading response body", "error", httpErr)
							}
							httpErr = resp.Body.Close()
							if httpErr != nil {
								slog.Warn("Error closing response body", "error", httpErr)
							}
							s.incSuccess()
						}
						wg.Done()
						<-guard
					}(i)
				}
				wg.Wait()
			}
		}
	}()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("{\"status\": \"ok\"}"))
	if err != nil {
		slog.Warn("Failed to write response", "error", err)
	}
}

func (s *SourceServer) stopHandler(w http.ResponseWriter, _ *http.Request) {
	s.notifyChan <- true
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("{\"status\": \"ok\"}"))
	if err != nil {
		slog.Warn("Failed to write response", "error", err)
	}
}

func (s *SourceServer) statusHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(fmt.Sprintf("{\"successCount\": \"%.0f\", \"failedCount\": \"%.0f\"}", s.successCount, s.failedCount)))
	if err != nil {
		slog.Warn("Failed to write response", "error", err)
	}
}

func (s *SourceServer) incSuccess() {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.successCount++
}

func (s *SourceServer) incFailed() {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.failedCount++
}
