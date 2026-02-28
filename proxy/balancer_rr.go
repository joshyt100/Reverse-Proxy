package proxy

import (
	"net/http"
	"net/url"
	"reverse-proxy/health"
	"sync/atomic"
)

type roundRobinBalancer struct {
	ups    []*url.URL
	h      *health.State
	rr     atomic.Uint64
	active []atomic.Int64
}

func newRoundRobinBalancer(ups []*url.URL, hs *health.State) *roundRobinBalancer {
	return &roundRobinBalancer{
		ups:    ups,
		h:      hs,
		active: make([]atomic.Int64, len(ups)),
	}
}

func (b *roundRobinBalancer) Pick(_ *http.Request) (*url.URL, func(), bool) {
	n := len(b.ups)
	if n == 0 {
		return nil, nil, false
	}

	start := int(b.rr.Add(1)-1) % n
	anyHealthy := b.h == nil || b.h.AnyHealthy()

	for k := 0; k < n; k++ {
		i := (start + k) % n
		if anyHealthy && b.h != nil && !b.h.IsHealthy(i) {
			continue
		}
		b.active[i].Add(1)
		return b.ups[i], func() { b.active[i].Add(-1) }, true
	}

	i := start
	b.active[i].Add(1)
	return b.ups[i], func() { b.active[i].Add(-1) }, true
}
