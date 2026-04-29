package proxy

import (
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

// LBAlgo identifies the load-balancing strategy used by the proxy.
type LBAlgo string

const (
	// LBRoundRobin selects upstreams in rotating order.
	LBRoundRobin LBAlgo = "rr"

	// LBLeastConn selects the healthy upstream with the fewest active connections.
	LBLeastConn LBAlgo = "lc"
)

// Options configures a Proxy instance.
type Options struct {
	// Upstreams is the list of backend servers the proxy can forward to.
	Upstreams []*url.URL

	// Transport optionally overrides the base HTTP transport.
	Transport *http.Transport

	// Algo controls which load-balancing strategy is used.
	Algo LBAlgo

	// Logger is used for structured proxy logs.
	Logger *slog.Logger

	// MetricsEnabled controls whether proxy metrics are recorded.
	MetricsEnabled bool

	// HealthPath is the path used for active health checks.
	HealthPath string

	// HealthInterval controls how often active health checks run.
	HealthInterval time.Duration

	// HealthTimeout controls how long a health check may take before failing.
	HealthTimeout time.Duration

	// PassiveFailCooldown controls how long an upstream is avoided after failure.
	PassiveFailCooldown time.Duration
}
