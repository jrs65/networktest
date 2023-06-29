package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

// Start and run the echo server at the given port
// This is a simple echo server handler that just returns whatever input it receives
func EchoServer(port int, logServer bool) {

	// The callback for the server
	echoCallback := func(w http.ResponseWriter, req *http.Request) {

		defer req.Body.Close()
		body, err := io.ReadAll(req.Body)

		if err != nil {
			return
		}

		fmt.Fprintf(w, "%s", string(body))

		if logServer {
			log.Printf("server: processing request from %s", req.RemoteAddr)
		}
	}

	http.HandleFunc("/", echoCallback)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
