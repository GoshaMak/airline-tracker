package postgres

import (
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

var poolCollectors sync.Map

type poolStats interface {
	TotalConns() int32
	AcquiredConns() int32
	IdleConns() int32
	ConstructingConns() int32
	MaxConns() int32
}

func registerPoolMetrics(pool *pgxpool.Pool) error {
	collector := newPoolCollector(func() poolStats {
		return pool.Stat()
	})
	if err := prometheus.Register(collector); err != nil {
		return err
	}
	poolCollectors.Store(pool, collector)
	return nil
}

type poolCollector struct {
	stats          func() poolStats
	connections    *prometheus.Desc
	maxConnections *prometheus.Desc
}

func newPoolCollector(stats func() poolStats) *poolCollector {
	return &poolCollector{
		stats: stats,
		connections: prometheus.NewDesc(
			"airline_tracker_postgres_pool_connections",
			"Current PostgreSQL pool connections by state; total includes acquired, idle and constructing.",
			[]string{"state"},
			nil,
		),
		maxConnections: prometheus.NewDesc(
			"airline_tracker_postgres_pool_max_connections",
			"Maximum PostgreSQL pool connections.",
			nil,
			nil,
		),
	}
}

func (c *poolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.connections
	ch <- c.maxConnections
}

func (c *poolCollector) Collect(ch chan<- prometheus.Metric) {
	stats := c.stats()
	for state, count := range map[string]int32{
		"total":        stats.TotalConns(),
		"acquired":     stats.AcquiredConns(),
		"idle":         stats.IdleConns(),
		"constructing": stats.ConstructingConns(),
	} {
		ch <- prometheus.MustNewConstMetric(c.connections, prometheus.GaugeValue, float64(count), state)
	}
	ch <- prometheus.MustNewConstMetric(c.maxConnections, prometheus.GaugeValue, float64(stats.MaxConns()))
}
