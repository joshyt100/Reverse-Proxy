package proxy

import (
	"math"
	"net/http"
	"net/url"
	"reverse-proxy/health"
	"sync/atomic"
)

type leastConnBalancer struct {
	ups    []*url.URL
	h      *health.State
	active []atomic.Int64
	rr     atomic.Uint64
}

func newLeastConnBalancer(ups []*url.URL, hs *health.State) *leastConnBalancer {
	return &leastConnBalancer{
		ups:    ups,
		h:      hs,
		active: make([]atomic.Int64, len(ups)),
	}
}

func (b *leastConnBalancer) Pick(_ *http.Request) (*url.URL, func(), bool) {
	n := len(b.ups)
	if n == 0 {
		return nil, nil, false
	}
	if n == 1 {
		b.active[0].Add(1)
		return b.ups[0], func() { b.active[0].Add(-1) }, true
	}

	start := int(b.rr.Add(1)-1) % n
	minIdx := -1
	minVal := int64(math.MaxInt64)
	anyHealthy := b.h == nil || b.h.AnyHealthy()

	for k := 0; k < n; k++ {
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

	if minIdx == -1 {
		minVal = int64(math.MaxInt64)
		for k := 0; k < n; k++ {
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
