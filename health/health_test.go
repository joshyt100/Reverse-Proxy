package health

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func mustURL(t *testing.T, raw string) *url.URL {
	t.Helper()

	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("failed to parse url %q: %v", raw, err)
	}

	return u
}

func TestCheckOne_Healthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Fatalf("expected /health, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ups := []*url.URL{mustURL(t, srv.URL)}
	state := NewState(ups, http.DefaultTransport, "/health", time.Second, time.Second, time.Second)

	if !state.checkOne(0) {
		t.Fatal("expected upstream to be healthy")
	}
}

func TestCheckOne_UnhealthyStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ups := []*url.URL{mustURL(t, srv.URL)}
	state := NewState(ups, http.DefaultTransport, "/health", time.Second, time.Second, time.Second)

	if state.checkOne(0) {
		t.Fatal("expected upstream to be unhealthy")
	}
}

func TestMarkPassiveFailure(t *testing.T) {
	ups := []*url.URL{mustURL(t, "http://example.com")}
	state := NewState(ups, http.DefaultTransport, "/health", time.Second, time.Second, 100*time.Millisecond)

	if !state.IsHealthy(0) {
		t.Fatal("expected upstream to start healthy")
	}

	state.MarkPassiveFailure(0)

	if state.IsHealthy(0) {
		t.Fatal("expected upstream to be temporarily unhealthy after passive failure")
	}

	time.Sleep(150 * time.Millisecond)

	if !state.IsHealthy(0) {
		t.Fatal("expected upstream to recover after passive cooldown")
	}
}

func TestAnyHealthy(t *testing.T) {
	ups := []*url.URL{
		mustURL(t, "http://one.example.com"),
		mustURL(t, "http://two.example.com"),
	}
	state := NewState(ups, http.DefaultTransport, "/health", time.Second, time.Second, time.Second)

	state.healthy[0].Store(false)
	state.healthy[1].Store(true)

	if !state.AnyHealthy() {
		t.Fatal("expected at least one healthy upstream")
	}

	state.healthy[1].Store(false)

	if state.AnyHealthy() {
		t.Fatal("expected no healthy upstreams")
	}
}

func TestJoinURLPath(t *testing.T) {
	got := joinURLPath("/base", "/health")
	want := "/base/health"

	if got != want {
		t.Fatalf("joinURLPath returned %q, want %q", got, want)
	}
}

// verify that if passive failure marked an upstream unhealthy, later successful active check makes it healthy immediately again.
func TestCheckOne_ClearsPassivePenaltyOnRecovery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ups := []*url.URL{mustURL(t, srv.URL)}
	state := NewState(ups, http.DefaultTransport, "/health", time.Second, time.Second, time.Second)

	state.MarkPassiveFailure(0)

	if state.IsHealthy(0) {
		t.Fatal("expected upstream to be unhealthy during passive cooldown")
	}

	if !state.checkOne(0) {
		t.Fatal("expected active health check to succeed")
	}

	if !state.IsHealthy(0) {
		t.Fatal("expected successful active health check to clear passive penalty")
	}
}

func TestCheckOne_UsesUpstreamBasePath(t *testing.T) {
	var gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	up := mustURL(t, srv.URL+"/api")
	ups := []*url.URL{up}
	state := NewState(ups, http.DefaultTransport, "/health", time.Second, time.Second, time.Second)

	if !state.checkOne(0) {
		t.Fatal("expected upstream to be healthy")
	}

	if gotPath != "/api/health" {
		t.Fatalf("expected request path %q, got %q", "/api/health", gotPath)
	}
}
