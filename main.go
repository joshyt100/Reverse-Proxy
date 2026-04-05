package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"reverse-proxy/config"
	"reverse-proxy/metrics"
	"reverse-proxy/proxy"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	metrics.Register()

	upstreams, err := proxy.ParseUpstreams(joinUpstreams(cfg.Upstreams))
	if err != nil {
		log.Fatalf("upstreams: %v", err)
	}

	algo := proxy.LBAlgo(cfg.Algo)

	if !cfg.Cleartext.Enabled && !cfg.TLS.Enabled {
		log.Fatal("at least one of cleartext or tls must be enabled")
	}

	p := proxy.New(proxy.Options{
		Upstreams:           upstreams,
		Algo:                algo,
		HealthPath:          "/",
		HealthInterval:      5 * time.Second,
		HealthTimeout:       2 * time.Second,
		PassiveFailCooldown: 10 * time.Second,
	})

	mux := http.NewServeMux()
	mux.Handle("/", p)
	mux.Handle("/metrics", promhttp.Handler())

	h2s := &http2.Server{
		MaxConcurrentStreams: 250,
	}
	h2cHandler := h2c.NewHandler(mux, h2s)
	if cfg.Cleartext.Enabled {
		clearSrv := &http.Server{
			Addr:    cfg.Cleartext.ListenAddr,
			Handler: h2cHandler,
		}

		if err := http2.ConfigureServer(clearSrv, h2s); err != nil {
			log.Fatalf("http2.ConfigureServer: %v", err)
		}

		go func() {
			log.Printf("proxy listening on %s (h2c + http/1.1)", cfg.Cleartext.ListenAddr)
			if err := clearSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}()
	}

	if cfg.TLS.Enabled {
		tlsSrv := &http.Server{
			Addr:    cfg.TLS.ListenAddr,
			Handler: mux,
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		}
		log.Printf("proxy listening on %s (tls + http/2)", cfg.TLS.ListenAddr)
		log.Fatal(tlsSrv.ListenAndServeTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile))
	} else {
		select {}
	}
}

func joinUpstreams(us []string) string {
	result := ""
	for i, u := range us {
		if i > 0 {
			result += ","
		}
		result += u
	}
	return result
}
