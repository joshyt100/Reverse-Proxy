package metrics

import "github.com/prometheus/client_golang/prometheus"

// RequestsTotal counts the total number of proxied requests.
// It is labeled by upstream and HTTP status code to allow
// per-backend and per-response analysis (e.g., error rates).
var (
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "proxy_requests_total",
			Help: "Total number of requests proxied.",
		},
		[]string{"upstream", "status_code"},
	)

	// RequestDuration measures the latency of proxied requests in seconds.
	// It uses Prometheus histogram buckets to support percentile queries
	// (e.g., p95, p99 latency) per upstream.
	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "proxy_request_duration_seconds",
			Help:    "Request latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"upstream"},
	)

	// ActiveConnections tracks the number of in-flight requests per upstream.
	// This is useful for monitoring load and for load balancing strategies
	// such as least-connections.
	ActiveConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "proxy_active_connections",
			Help: "Number of active connections per upstream.",
		},
		[]string{"upstream"},
	)

	// UpstreamHealthy indicates the health status of each upstream.
	// A value of 1 means healthy, and 0 means unhealthy.
	// This is updated by active/passive health checks.
	UpstreamHealthy = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "proxy_upstream_healthy",
			Help: "1 if upstream is healthy, 0 if not.",
		},
		[]string{"upstream"},
	)
)

// Register registers all proxy metrics with the default Prometheus registry.
// This must be called before exposing the /metrics endpoint so that all
// defined metrics are included in the output.
func Register() {
	prometheus.MustRegister(
		RequestsTotal,
		RequestDuration,
		ActiveConnections,
		UpstreamHealthy,
	)
}
