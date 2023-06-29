# networktest

A simple tool for network debugging.

Each running instance contains an HTTP client and server running on a given port. The
client will periodically connect to other running server instances and log whether the
attempt was successful and how long it took. The server simply echos whatever it
receives.

To use this start a copy up on every host that will be the destination of a test, and
specify the remote hosts that you would like it to connect to for testing.
```
$ networktest host1 host2 host3
```

There are several ways to record the results of these tests:
- Logging the results to stderr, either a summary or all results (with `-v`)
- Exporting the results as prometheus metrics (`--prometheus`)
- Writing the results to a file (`-o`)

For more details, see the help.
```
$ networktest --help
Poll servers on other instances to test the network status.

Usage: main [--interval INTERVAL] [--port PORT] [--resolve] [--verbose] [--log-server] [--log-interval LOG-INTERVAL] [--output OUTPUT] [--prometheus] [--prom-port PROM-PORT] [HOSTNAME [HOSTNAME ...]]

Positional arguments:
  HOSTNAME               name of each host to test

Options:
  --interval INTERVAL, -t INTERVAL
                         the interval between polling hosts. [default: 5s]
  --port PORT, -p PORT   the port to operate on. [default: 8090]
  --resolve, -r          resolve the hostnames to IP addresses at the start.
  --verbose, -v          log every connection to attempt.
  --log-server           log connections to the server
  --log-interval LOG-INTERVAL
                         the interval between writing summaries to the log. [default: 30s]
  --output OUTPUT, -o OUTPUT
                         write the test results to the given file
  --prometheus           export the test results over prometheus
  --prom-port PROM-PORT
                         the port on which to export prometheus metrics. [default: 8091]
  --help, -h             display this help and exit
```