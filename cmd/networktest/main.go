package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type StatusCode int

const (
	StatusSuccess StatusCode = iota
	StatusTimeout
	StatusFail
)

var Hosts map[string]string

// Hold the results of a network test
type TestStatus struct {

	// The source hostname
	Source string

	// The destination hostname
	Dest string

	// The test status
	Status StatusCode

	// The amount of time elapsed
	Elapsed float32

	// The actual time the test finished
	Completed time.Time
}

// A simple echo server handler. Just returns whatever input it receives
func echo(w http.ResponseWriter, req *http.Request) {

	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)

	if err != nil {
		return
	}

	fmt.Fprintf(w, string(body))
}

// Connect to the given host:port and return the results on the channel
func check_host(host string, port int, ch chan<- TestStatus) {

	// Try to get the host from a pre-resolved map, otherwise just use the given name
	// and hope it works fine
	resolved_host, ok := Hosts[host]
	if !ok {
		resolved_host = host
	}

	// Dummy operation for the call
	url := fmt.Sprintf("http://%s:%d/", resolved_host, port)
	time.Sleep(800 * time.Millisecond)
	fmt.Println(url, time.Now())

	// Connect to the server, force a new connection each time
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConnsPerHost = -1
	startTime := time.Now()
	client := &http.Client{Timeout: 10 * time.Second, Transport: t}
	_, err := client.Get(url)
	endTime := time.Now()

	elapsed := float32(endTime.Sub(startTime).Seconds())

	var status StatusCode
	if err, ok := err.(net.Error); ok && err.Timeout() {
		// A timeout error occurred
		status = StatusTimeout
	} else if err != nil {
		// This was an error, but not a timeout
		status = StatusFail
	} else {
		status = StatusSuccess
	}

	// Create the status entry and return it
	s := TestStatus{"me", host, status, elapsed, endTime}
	ch <- s
}

// Wait for results on the channel and log them
func log_results(ch <-chan TestStatus) {

	for {
		select {
		case status := <-ch:
			//fmt.Println(status.Source, status.Dest, status.Status, status.Elapsed, status.Completed)
			log.Print(
				status.Source, status.Dest, status.Status, status.Elapsed,
				status.Completed,
			)
		}
	}
}

func check(hosts []string, port int, interval_ms int) {

	ch := make(chan TestStatus)
	go log_results(ch)

	for {
		for _, host := range hosts {
			go check_host(host, port, ch)
		}
		time.Sleep(time.Duration(interval_ms) * time.Millisecond)
	}
}

func myUsage() {
	fmt.Printf("Usage: %s [OPTIONS] HOSTNAME1, ...\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {

	portPtr := flag.Int("port", 8090, "The port to listen on.")
	resolvePtr := flag.Bool(
		"resolve", false, "Resolve the hostnames to IPs in advance.",
	)

	flag.Usage = myUsage
	flag.Parse()

	Hosts = make(map[string]string)

	hostsToUse := flag.Args()

	for _, host := range hostsToUse {

		addresses, err := net.LookupHost(host)

		if err != nil {
			log.Fatal(err)
		} else if len(addresses) == 0 {
			log.Fatalf("No address for host %s", host)
		}

		if *resolvePtr {
			Hosts[host] = addresses[0]
			//Hosts["localhost"] = "127.0.0.1"
		}
	}

	//hosts := []string{"localhost", "localhost", "localhost"}

	go check(hostsToUse, *portPtr, 1000)

	fmt.Println("Hello world")

	http.HandleFunc("/", echo)
	http.ListenAndServe(fmt.Sprintf(":%d", *portPtr), nil)
}
