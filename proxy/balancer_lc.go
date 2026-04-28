package proxy

import (
	"math"
	"net/http"
	"net/url"
	"reverse-proxy/health"
	"sync/atomic"
)

// leastConnBalancer distributes requests to whichever upstream currently has
// the fewest active connections. When multiple upstreams are tied, the one
// earliest in a round-robin rotation is chosen so load spreads evenly at rest.
type leastConnBalancer struct {
	ups    []*url.URL
	h      *health.State
	active []atomic.Int64 // tracks in-flight request count per upstream
	rr     atomic.Uint64  // incremented each Pick to vary the scan start point
}

// newLeastConnBalancer constructs a leastConnBalancer for the given upstreams.
// All active counters start at zero.
func newLeastConnBalancer(ups []*url.URL, hs *health.State) *leastConnBalancer {
	return &leastConnBalancer{
		ups:    ups,
		h:      hs,
		active: make([]atomic.Int64, len(ups)),
	}
}

// Pick selects the healthy upstream with the fewest active connections and
// increments its counter. It returns the chosen URL and a done function that
// the caller must invoke when the request finishes — this decrements the
// counter so future picks reflect the completed request.
//
// The scan begins at a round-robin offset rather than always index 0, so that
// upstreams with equal connection counts are chosen in rotation rather than
// always favouring the first one.
//
// If all upstreams are unhealthy, Pick falls back to the least-loaded upstream
// regardless of health rather than refusing traffic entirely. If there are no
// upstreams at all, Pick returns false.
func (b *leastConnBalancer) Pick(_ *http.Request) (*url.URL, func(), bool) {
	n := len(b.ups)
	if n == 0 {
		return nil, nil, false
	}

	// Fast path: only one upstream, no selection needed.
	if n == 1 {
		b.active[0].Add(1)
		return b.ups[0], func() { b.active[0].Add(-1) }, true
	}

	// Vary the start index each call so ties are broken in rotation.
	start := int(b.rr.Add(1)-1) % n

	// First pass: find the least-loaded healthy upstream.
	minIdx := -1
	minVal := int64(math.MaxInt64)
	anyHealthy := b.h == nil || b.h.AnyHealthy()
	for k := range n {
		i := (start + k) % n
		if anyHealthy && b.h != nil && !b.h.IsHealthy(i) {
			continue
		}
		v := b.active[i].Load()
		if v < minVal {
			minVal = v
			minIdx = i
		}
	}

	// Second pass: if no healthy upstream was found, fall back to the
	// least-loaded upstream unconditionally rather than dropping the request.
	if minIdx == -1 {
		minVal = int64(math.MaxInt64)
		for k := range n {
			i := (start + k) % n
			v := b.active[i].Load()
			if v < minVal {
				minVal = v
				minIdx = i
			}
		}
	}

	b.active[minIdx].Add(1)
	return b.ups[minIdx], func() { b.active[minIdx].Add(-1) }, true
}
