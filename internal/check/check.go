package check

import (
	"fmt"
	"log"
	"net"
	"net/http"
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

// Connect to the given host:port and return the results on the channel
func checkHost(name string, host string, port int, handlerChannels [](chan<- TestStatus)) {

	// Try to get the host from a pre-resolved map, otherwise just use the given name
	// and hope it works fine
	resolved_host, ok := Hosts[host]
	if !ok {
		resolved_host = host
	}

	// Dummy operation for the call
	url := fmt.Sprintf("http://%s:%d/", resolved_host, port)
	// time.Sleep(800 * time.Millisecond)
	// fmt.Println(url, time.Now())

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
	s := TestStatus{"me", name, status, elapsed, endTime}

	// Push the status to all listening handlers
	for _, channel := range handlerChannels {
		channel <- s
	}
}

func CheckHosts(
	hosts []string, port int, interval_ms int, resolve bool,
	handlers []HandlerCallback,
) {

	urlHosts := make([]string, len(hosts))

	for i, host := range hosts {

		var address string
		if host == "localhost" {
			address = "127.0.0.1"
		} else {
			addresses, err := net.LookupHost(host)

			if err != nil {
				log.Fatal(err)
			} else if len(addresses) == 0 {
				log.Fatalf("No address for host %s", host)
			}
			address = addresses[0]
		}

		if resolve {
			urlHosts[i] = address
		} else {
			urlHosts[i] = host
		}
	}

	// Create a channel to communicate with each handler and start up their go routines
	handlerChannels := make([](chan<- TestStatus), len(handlers))
	for i, handler := range handlers {
		c := make(chan TestStatus)
		handlerChannels[i] = c
		go handler(c)
	}

	for {
		for i, host := range hosts {
			go checkHost(host, urlHosts[i], port, handlerChannels)
		}
		time.Sleep(time.Duration(interval_ms) * time.Millisecond)
	}
}
