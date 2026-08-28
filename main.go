package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/tjololo/websoaker/cmd"

	flag "github.com/spf13/pflag"
)

func main() {
	var port string
	var concurrency int
	var maxCons int
	var host string
	var basePath string
	var logLevel int
	flag.StringVar(&port, "port", "8080", "Port to listen on")
	flag.IntVar(&concurrency, "concurrency", 10, "Number of concurrent requests to make")
	flag.StringVar(&host, "address", "http://localhost:8080", "Websoaker address for the sink server")
	flag.StringVar(&basePath, "basePath", "", "Base path for the sink ping endpoint")
	flag.IntVar(&maxCons, "maxCons", 1000, "Max connections per host")
	flag.IntVar(&logLevel, "logLevel", 2, "Log level (1: Error, 2: Warning, 3: Info, 4: Debug)")
	flag.Parse()
	args := os.Args
	setupLogger(logLevel)
	if len(args) < 2 {
		slog.Error("Usage: websoaker source|sink")
		os.Exit(1)
	}
	switch args[1] {
	case "source":
		if basePath != "" && !strings.HasPrefix(basePath, "/") {
			basePath = fmt.Sprintf("/%s", basePath)
		}
		sourceServer := cmd.NewSourceServer(host, basePath, concurrency, maxCons)
		sourceServer.StartSourceServer(port)
	case "sink":
		sinkServer := cmd.SinkServer{}
		sinkServer.StartSinkServer(port)
	default:
		slog.Error("Usage: websoaker source|sink")
		os.Exit(1)
	}
}

func setupLogger(logLevel int) {
	loglevel := slog.LevelError
	switch logLevel {
	case 1:
		loglevel = slog.LevelError
	case 2:
		loglevel = slog.LevelWarn
	case 3:
		loglevel = slog.LevelInfo
	case 4:
		loglevel = slog.LevelDebug
	default:
		loglevel = slog.LevelError
	}
	opts := &slog.HandlerOptions{
		Level: loglevel,
	}
	consoleHandler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(consoleHandler)
	slog.SetDefault(logger)
}
