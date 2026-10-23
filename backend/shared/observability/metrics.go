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

var jobs = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "airline_tracker_jobs_total",
	Help: "Background operations by worker and bounded result (success, retry, failed, duplicate).",
}, []string{"service", "result"})

var jobDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "airline_tracker_job_duration_seconds",
	Help:    "Background operation duration by worker and result.",
	Buckets: prometheus.DefBuckets,
}, []string{"service", "result"})

var queuePending = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "airline_tracker_queue_pending",
	Help: "Pending durable jobs, including jobs waiting for their next attempt.",
}, []string{"service", "queue"})

var queueOldest = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "airline_tracker_queue_oldest_seconds",
	Help: "Age of the oldest pending durable job.",
}, []string{"service", "queue"})

func ObserveJob(service, result string, elapsed time.Duration) {
	jobs.WithLabelValues(service, result).Inc()
	jobDuration.WithLabelValues(service, result).Observe(elapsed.Seconds())
}

func SetQueue(service, queue string, pendingAmt, oldestSeconds float64) {
	queuePending.WithLabelValues(service, queue).Set(pendingAmt)
	queueOldest.WithLabelValues(service, queue).Set(oldestSeconds)
}

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
