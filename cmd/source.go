package cmd

import (
	"fmt"
	"github.com/tjololo/websoaker/internal/server"
	"log"
	"net/http"
	"sync"
	"time"
)

type SourceServer struct {
	running      bool
	parallelism  int
	notifyChan   chan bool
	soakAddr     string
	mux          sync.Mutex
	successCount float64
	failedCount  float64
}

func NewSourceServer(soakAddr string, parallelism int) *SourceServer {
	return &SourceServer{
		running:     false,
		parallelism: parallelism,
		notifyChan:  make(chan bool),
		soakAddr:    soakAddr,
	}
}

func (s *SourceServer) StartSourceServer(port string) {
	log.Printf("Starting source server listening on port %s with %d concurrent", port, s.parallelism)
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
			log.Printf("Failed to write response %v", err)
		}
		return
	}
	s.running = true
	go func() {
		client := &http.Client{
			Timeout: 5 * time.Second,
		}
		for {
			select {
			case <-s.notifyChan:
				log.Println("Stop notification received")
				return
			default:
				guard := make(chan struct{}, s.parallelism)
				wg := &sync.WaitGroup{}
				for i := 0; i < s.parallelism; i++ {
					wg.Add(1)
					guard <- struct{}{}
					go func(n int) {
						resp, err := client.Get(s.soakAddr + "/ping")
						if err != nil {
							log.Printf("Error making request: %s", err)
							s.incFailed()
						} else {
							closeErr := resp.Body.Close()
							if closeErr != nil {
								log.Printf("Error closing response body: %s", closeErr)
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
		log.Printf("Failed to write response %v", err)
	}
}

func (s *SourceServer) stopHandler(w http.ResponseWriter, r *http.Request) {
	s.notifyChan <- true
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("{\"status\": \"ok\"}"))
	if err != nil {
		log.Printf("Failed to write response %v", err)
	}
}

func (s *SourceServer) statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(fmt.Sprintf("{\"successCount\": \"%.0f\", \"failedCount\": \"%.0f\"}", s.successCount, s.failedCount)))
	if err != nil {
		log.Printf("Failed to write response %v", err)
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
