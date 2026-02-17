package proxy

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

type Options struct {
	Upstreams []*url.URL
	Client    *http.Client
}

type Proxy struct {
	upstreams []*url.URL
	client    *http.Client
	rr        uint64
}

func New(opts Options) *Proxy {
	c := opts.Client
	if c == nil {
		c = &http.Client{Timeout: 60 * time.Second}
	}
	return &Proxy{
		upstreams: opts.Upstreams,
		client:    c,
	}
}

func ParseUpstreams(csv string) ([]*url.URL, error) {
	parts := strings.Split(csv, ",")
	out := make([]*url.URL, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		u, err := url.Parse(p)
		if err != nil {
			return nil, fmt.Errorf("invalid upstream %q: %w", p, err)
		}
		if u.Scheme == "" || u.Host == "" {
			return nil, fmt.Errorf("upstream must include scheme and host, got %q", p)
		}
		u.Path = strings.TrimRight(u.Path, "/")
		out = append(out, u)
	}
	return out, nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	up := p.pickUpstream()

	outReq, err := p.buildUpstreamRequest(r, up)
	if err != nil {
		http.Error(w, "bad gateway", http.StatusBadGateway)
		return
	}

	resp, err := p.client.Do(outReq)
	if err != nil {
		http.Error(w, "bad gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	copyHeaders(w.Header(), resp.Header)
	removeHopByHopHeaders(w.Header())

	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (p *Proxy) pickUpstream() *url.URL {
	i := int(atomic.AddUint64(&p.rr, 1)-1) % len(p.upstreams)
	return p.upstreams[i]
}

func (p *Proxy) buildUpstreamRequest(in *http.Request, up *url.URL) (*http.Request, error) {
	target := *up
	target.Path = joinURLPath(up.Path, in.URL.Path)
	target.RawQuery = in.URL.RawQuery

	outReq, err := http.NewRequestWithContext(in.Context(), in.Method, target.String(), in.Body)
	if err != nil {
		return nil, err
	}

	copyHeaders(outReq.Header, in.Header)
	removeHopByHopHeaders(outReq.Header)

	outReq.Host = up.Host

	addXForwardedFor(outReq.Header, in)
	addXForwardedProto(outReq.Header, in)
	addXForwardedHost(outReq.Header, in)

	return outReq, nil
}

func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		k = textproto.CanonicalMIMEHeaderKey(k)
		dst.Del(k)
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func removeHopByHopHeaders(h http.Header) {
	if c := h.Get("Connection"); c != "" {
		for _, f := range strings.Split(c, ",") {
			if name := strings.TrimSpace(f); name != "" {
				h.Del(name)
			}
		}
	}

	h.Del("Connection")
	h.Del("Proxy-Connection")
	h.Del("Keep-Alive")
	h.Del("Proxy-Authenticate")
	h.Del("Proxy-Authorization")
	h.Del("TE")
	h.Del("Trailer")
	h.Del("Transfer-Encoding")
	h.Del("Upgrade")
}

func addXForwardedFor(h http.Header, in *http.Request) {
	ip := clientIP(in)
	if ip == "" {
		return
	}
	if prior := h.Get("X-Forwarded-For"); prior != "" {
		h.Set("X-Forwarded-For", prior+", "+ip)
	} else {
		h.Set("X-Forwarded-For", ip)
	}
}

func addXForwardedProto(h http.Header, in *http.Request) {
	proto := "http"
	if in.TLS != nil {
		proto = "https"
	}
	h.Set("X-Forwarded-Proto", proto)
}

func addXForwardedHost(h http.Header, in *http.Request) {
	if in.Host != "" {
		h.Set("X-Forwarded-Host", in.Host)
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	if net.ParseIP(r.RemoteAddr) != nil {
		return r.RemoteAddr
	}
	return ""
}

func joinURLPath(a, b string) string {
	switch {
	case a == "" || a == "/":
		return cleanPath(b)
	case b == "" || b == "/":
		return cleanPath(a)
	default:
		return cleanPath(strings.TrimRight(a, "/") + "/" + strings.TrimLeft(b, "/"))
	}
}

func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}
