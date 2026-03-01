package main

import (
	"flag"
	"log"
	"net/http"
	"reverse-proxy/config"
	"reverse-proxy/proxy"
	"strings"
	"time"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	// flags
	listen := flag.String("listen", cfg.ListenAddr, "listen address")
	upstreamsCSV := flag.String("upstreams", "", "comma-separated upstream URLs (overrides config)")
	flag.Parse()

	cfg.ListenAddr = *listen
	if *upstreamsCSV != "" {
		cfg.Upstreams = strings.Split(*upstreamsCSV, ",")
	}

	upstreams, err := proxy.ParseUpstreams(strings.Join(cfg.Upstreams, ","))
	if err != nil {
		log.Fatal(err)
	}
	if len(upstreams) == 0 {
		log.Fatal("no upstreams provided")
	}

	p := proxy.New(proxy.Options{
		Upstreams: upstreams,
	})

	srv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           p,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("listening on %s -> %v", cfg.ListenAddr, upstreams)
	log.Fatal(srv.ListenAndServe())
}
