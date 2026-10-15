package observability

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var requests = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "airline_tracker_http_requests_total",
	Help: "Completed API requests by method, route template and status.",
}, []string{"method", "route", "status"})

var duration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "airline_tracker_http_request_duration_seconds",
	Help:    "API request duration in seconds.",
	Buckets: prometheus.DefBuckets,
}, []string{"method", "route", "status"})

func ObserveHTTP(method, route string, status int, elapsed time.Duration) {
	switch method {
	case "GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "CONNECT", "TRACE":
	default:
		method = "OTHER"
	}
	if route == "" {
		route = "unmatched"
	}
	code := strconv.Itoa(status)
	requests.WithLabelValues(method, route, code).Inc()
	duration.WithLabelValues(method, route, code).Observe(elapsed.Seconds())
}

func Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	server := &http.Server{
		Addr:              ":9090",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
	}
	shutdownDone := make(chan struct{})
	stop := context.AfterFunc(ctx, func() {
		defer close(shutdownDone)
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			slog.Error("metrics shutdown failed", "err", err)
		}
	})
	defer func() {
		if !stop() {
			<-shutdownDone
		}
	}()
	slog.Info("metrics server started", "address", server.Addr)
	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
