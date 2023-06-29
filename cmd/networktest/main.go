package main

import (
	"time"

	arg "github.com/alexflint/go-arg"

	"github.com/jrs65/networktest/internal/check"
	"github.com/jrs65/networktest/internal/server"
)

// Define the command line arguments we can accept
type cliArgs struct {
	Interval time.Duration `arg:"-t,--interval" default:"5s" help:"the interval between polling hosts."`
	Port     int           `arg:"-p,--port" help:"the port to operate on." default:"8090"`
	Resolve  bool          `arg:"-r,--resolve" help:"resolve the hostnames to IP addresses at the start."`

	Verbose     bool          `arg:"-v,--verbose" help:"log every connection to attempt."`
	LogServer   bool          `arg:"--log-server" help:"log connections to the server"`
	LogInterval time.Duration `arg:"--log-interval" default:"30s" help:"the interval between writing summaries to the log."`

	OutputPath string `arg:"-o,--output" help:"write the test results to the given file"`

	Prometheus     bool `help:"export the test results over prometheus"`
	PrometheusPort int  `arg:"--prom-port" default:"8091" help:"the port on which to export prometheus metrics."`

	Hosts []string `arg:"positional" help:"name of each host to test" placeholder:"HOSTNAME"`
}

func (cliArgs) Description() string {
	return "Poll servers on other instances to test the network status.\n"
}

func main() {

	// TODO: improve file argument handling

	var args cliArgs
	arg.MustParse(&args)

	handlers := make([]check.HandlerCallback, 0, 10)

	// Log handlers
	handlers = append(handlers, check.LogSummary(args.LogInterval))
	if args.Verbose {
		handlers = append(handlers, check.LogVerbose())
	}

	// File output handlers
	if args.OutputPath != "" {
		handlers = append(handlers, check.FileWriter(args.OutputPath))
	}

	// // Prometheus handler
	if args.Prometheus {
		handlers = append(handlers, check.Prometheus(args.PrometheusPort))
	}

	// Start the client checking all the hosts in a go routine
	go check.CheckHosts(args.Hosts, args.Port, args.Interval, args.Resolve, handlers)

	// Start up the server
	server.EchoServer(args.Port, args.LogServer)
}
