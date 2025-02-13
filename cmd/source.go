package cmd

import (
	"fmt"
	"github.com/tjololo/websoaker/internal/server"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
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
	log.Printf("Starting source server listening on port %s with %d concurrent", port, s.concurrency)
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
				log.Println("Stop notification received")
				return
			default:
				guard := make(chan struct{}, s.concurrency)
				wg := &sync.WaitGroup{}
				for i := 0; i < s.concurrency; i++ {
					wg.Add(1)
					guard <- struct{}{}
					go func(n int) {
						resp, err := client.Get(fmt.Sprintf("%s/%s/ping", s.soakHost, s.basePath))
						if err != nil {
							log.Printf("Error making request: %s", err)
							s.incFailed()
						} else {
							_, httpErr := io.Copy(io.Discard, resp.Body)
							if httpErr != nil {
								log.Printf("Error reading response body: %s", httpErr)
							}
							httpErr = resp.Body.Close()
							if httpErr != nil {
								log.Printf("Error closing response body: %s", httpErr)
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

func (s *SourceServer) stopHandler(w http.ResponseWriter, _ *http.Request) {
	s.notifyChan <- true
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("{\"status\": \"ok\"}"))
	if err != nil {
		log.Printf("Failed to write response %v", err)
	}
}

func (s *SourceServer) statusHandler(w http.ResponseWriter, _ *http.Request) {
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
