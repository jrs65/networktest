package check

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HandlerCallback func(<-chan TestStatus)

// Wait for results on the channel and log them
func LogVerbose() HandlerCallback {
	return func(ch <-chan TestStatus) {

		for status := range ch {

			var statusString string

			if status.Status == StatusSuccess {
				statusString = "SUCCESS"
			} else {
				statusString = "FAILURE"
			}
			log.Printf(
				"%10s <-> %-10s %s in %.1f ms",
				status.Source, status.Dest, statusString, 1000*status.Elapsed,
			)
		}
	}
}

// Wait for results on the channel and write an entry to a file
func FileWriter(filename string) HandlerCallback {

	return func(ch <-chan TestStatus) {

		file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)

		if err != nil {
			log.Fatalf("Count not open file %s", filename)
		}

		defer file.Close()

		for status := range ch {
			fmt.Fprintf(
				file,
				"%s %15s %15s %15d %10.1f\n",
				status.Completed.Format(time.RFC3339),
				status.Source, status.Dest, status.Status, 1000*status.Elapsed,
			)
		}
	}
}

// Write a summary out to the log every `interval` seconds
func LogSummary(interval float64) HandlerCallback {

	failCounts := make(map[string]int)
	totalCounts := make(map[string]int)

	return func(ch <-chan TestStatus) {

		lastTime := time.Now()

		for status := range ch {
			totalCounts[status.Dest] += 1

			if status.Status != StatusSuccess {
				failCounts[status.Dest] += 1
			}

			// Print out all the counts if we're at a multiple of 10 tests
			if time.Since(lastTime).Seconds() > interval {
				for dest := range totalCounts {
					log.Printf(
						"%s %d/%d failures",
						dest, failCounts[dest], totalCounts[dest],
					)
				}

				lastTime = time.Now()
			}
		}
	}
}

func Prometheus(port int) HandlerCallback {

	failCounter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "networktest_connection_fail_total",
			Help: "Total number of failed connections attempts.",
		},
		[]string{"source", "dest"},
	)

	totalCounter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "networktest_connection_attempts_total",
			Help: "Total number of connections attempts.",
		},
		[]string{"source", "dest"},
	)

	return func(ch <-chan TestStatus) {

		http.Handle("/metrics", promhttp.Handler())
		go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

		for status := range ch {

			totalCounter.WithLabelValues(status.Source, status.Dest).Inc()

			if status.Status != StatusSuccess {
				failCounter.WithLabelValues(status.Source, status.Dest).Inc()
			}
		}
	}
}
