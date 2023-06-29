package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jrs65/networktest/internal/check"
	"github.com/jrs65/networktest/internal/server"
)

func myUsage() {
	fmt.Printf("Usage: %s [OPTIONS] HOSTNAME1, ...\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {

	// TODO: improve CLI parsing
	// TODO: add a verbose option instead of log
	// TODO: improve file argument handling
	// TODO: add server logging
	// TODO: add prometheus logging handler
	// TODO: improving config handling
	// TODO: trying building on tubular

	portPtr := flag.Int("port", 8090, "The port to listen on.")
	resolvePtr := flag.Bool(
		"resolve", false, "Resolve the hostnames to IPs in advance.",
	)
	outPtr := flag.String("out", "", "A file to write detailed results to.")
	logPtr := flag.Bool("log", false, "Log each connection attempt.")

	flag.Usage = myUsage
	flag.Parse()

	hostsToUse := flag.Args()

	//hosts := []string{"localhost", "localhost", "localhost"}

	handlers := make([]check.HandlerCallback, 0, 10)

	if *logPtr {
		handlers = append(handlers, check.LogVerbose())
	}

	if *outPtr != "" {
		handlers = append(handlers, check.FileWriter(*outPtr))
	}
	handlers = append(handlers, check.LogSummary(10.0))
	handlers = append(handlers, check.Prometheus(1781))

	go check.CheckHosts(hostsToUse, *portPtr, 1000, *resolvePtr, handlers)

	//fmt.Println("Hello world")

	server.EchoServer(*portPtr)
}
