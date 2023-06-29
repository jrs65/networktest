package server

import (
	"fmt"
	"io"
	"net/http"
)

// The callback for the server
func echoCallback(w http.ResponseWriter, req *http.Request) {

	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)

	if err != nil {
		return
	}

	fmt.Fprintf(w, "%s", string(body))
}

// Start and run the echo server at the given port
// This is a simple echo server handler that just returns whatever input it receives
func EchoServer(port int) {
	http.HandleFunc("/", echoCallback)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
