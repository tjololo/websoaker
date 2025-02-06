package main

import (
	"github.com/tjololo/websoaker/cmd"
	"log"
	"os"

	flag "github.com/spf13/pflag"
)

func main() {
	var port string
	var concurrency int
	var maxCons int
	var address string
	var basePath string
	flag.StringVar(&port, "port", "8080", "Port to listen on")
	flag.IntVar(&concurrency, "concurrency", 10, "Number of concurrent requests to make")
	flag.StringVar(&address, "address", "http://localhost:8080", "Websoaker address for the sink server")
	flag.StringVar(&basePath, "basePath", "", "Base path for the sink ping endpoint")
	flag.IntVar(&maxCons, "maxCons", 1000, "Max connections per host")
	flag.Parse()
	args := os.Args
	if len(args) < 2 {
		log.Fatal("Usage: websoaker source|sink")
	}
	switch args[1] {
	case "source":
		sourceServer := cmd.NewSourceServer(address, basePath, concurrency, maxCons)
		sourceServer.StartSourceServer(port)
	case "sink":
		sinkServer := cmd.SinkServer{}
		sinkServer.StartSinkServer(port)
	default:
		log.Fatal("Usage: websoaker source|sink")
	}
}
