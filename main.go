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
	var address string
	var rootPath string
	flag.StringVar(&port, "port", "8080", "Port to listen on")
	flag.IntVar(&concurrency, "concurrency", 10, "Number of concurrent requests to make")
	flag.StringVar(&address, "address", "http://localhost:8080", "Websoaker address for the sink server")
	flag.StringVar(&rootPath, "rootPath", "", "Root path for the sink ping endpoint")
	flag.Parse()
	args := os.Args
	if len(args) < 2 {
		log.Fatal("Usage: websoaker source|sink")
	}
	switch args[1] {
	case "source":
		sourceServer := cmd.NewSourceServer(address, concurrency)
		sourceServer.StartSourceServer(port)
	case "sink":
		sinkServer := cmd.SinkServer{}
		sinkServer.StartSinkServer(port)
	default:
		log.Fatal("Usage: websoaker source|sink")
	}
}
