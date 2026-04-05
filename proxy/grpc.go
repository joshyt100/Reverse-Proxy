package proxy

import (
	"crypto/tls"
	"net"
	"net/http"
	"strings"

	"golang.org/x/net/http2"
)

// func isH2CUpstream(uScheme string) bool {
// 	return strings.EqualFold(uScheme, "h2c")
// }

// newH2CTransport returns an http2.Transport that dials plain TCP.
// DialTLS is called even for cleartext h2c connections; we simply
// return a raw TCP conn and let the HTTP/2 framer take over.
func newH2CTransport() *http2.Transport {
	return &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(network, addr string, _ *tls.Config) (net.Conn, error) {
			return net.Dial(network, addr)
		},
	}
}

// preserveOrDropTE keeps TE: trailers (required by gRPC / HTTP/2) and
// drops any other TE value, which is a hop-by-hop concern.
func preserveOrDropTE(h http.Header) {
	te := h.Get("TE")
	if te == "" {
		return
	}
	if strings.EqualFold(strings.TrimSpace(te), "trailers") {
		h.Set("TE", "trailers")
		return
	}
	h.Del("TE")
}
