package check

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type StatusCode int

// Define the test status
const (
	StatusSuccess StatusCode = iota // Everything worked
	StatusTimeout                   // There was a timeout
	StatusFail                      // Some other failure happened
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
	Elapsed time.Duration

	// The actual time the test finished
	StartTime time.Time
}

// Where should we connect to
type TestHost struct {

	// The name of the destination (not necessarily the actual hostname)
	Name string

	// The actual host to connect to. This could be a hostname, or an IP address
	Host string

	// The port to connect to
	Port int
}

// Construct the URL to connect to
func (t TestHost) Url() string {
	return fmt.Sprintf("http://%s:%d/", t.Host, t.Port)
}

// Connect to the given destination and pass the results to the handlers
// source is the name to use for the origin (i.e this machine)
func checkHost(dest TestHost, source string, handlerChannels [](chan<- TestStatus)) {

	// Connect to the server, force a new connection each time
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConnsPerHost = -1
	startTime := time.Now()
	client := &http.Client{Timeout: 10 * time.Second, Transport: t}
	_, err := client.Get(dest.Url())
	endTime := time.Now()

	elapsed := endTime.Sub(startTime)

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
	s := TestStatus{source, dest.Name, status, elapsed, startTime}

	// Push the status to all listening handlers
	for _, channel := range handlerChannels {
		channel <- s
	}
}

// Check that we can connect to the given list of `hosts` on `port`.
// If `resolve` then try to pre-resolve the names to IP addresses to avoid DNS queries
// on every connection attempt
// The results of each test are passed to each element of `handlers` via a channel.
func CheckHosts(
	hosts []string, port int, interval time.Duration, resolve bool,
	handlers []HandlerCallback,
) {

	testHosts := make([]TestHost, len(hosts))

	for i, host := range hosts {

		addresses, err := net.LookupHost(host)

		if err != nil {
			log.Fatal(err)
		} else if len(addresses) == 0 {
			log.Fatalf("No address for host %s", host)
		}

		testHosts[i].Name = host
		testHosts[i].Port = port

		if resolve {
			log.Printf("resolved host %s to %s", host, addresses[0])
			testHosts[i].Host = addresses[0]
		} else {
			testHosts[i].Host = host
		}
	}

	// Create a channel to communicate with each handler and start up their go routines
	handlerChannels := make([](chan<- TestStatus), len(handlers))
	for i, handler := range handlers {
		c := make(chan TestStatus)
		handlerChannels[i] = c
		go handler(c)
	}

	// Get the local hostname to label the test origin
	me, err := os.Hostname()
	if err != nil {
		me = "me"
	}

	for {
		for _, t := range testHosts {
			go checkHost(t, me, handlerChannels)
		}
		time.Sleep(interval)
	}
}
