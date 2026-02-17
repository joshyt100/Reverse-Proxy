package main

import (
	"flag"
	"log"
	"net/http"
	"reverse-proxy/proxy"
	"time"
)

func main() {
	var listenAddr string
	var upstreamsCSV string

	flag.StringVar(&listenAddr, "listen", ":8080", "listen address")
	flag.StringVar(&upstreamsCSV, "upstreams", "http://localhost:9000", "comma-separated upstream base URLs")
	flag.Parse()

	upstreams, err := proxy.ParseUpstreams(upstreamsCSV)
	if err != nil {
		log.Fatal(err)
	}
	if len(upstreams) == 0 {
		log.Fatal("no upstreams provided")
	}

	p := proxy.New(proxy.Options{
		Upstreams: upstreams,
		Client: &http.Client{
			Timeout: 60 * time.Second,
		},
	})

	srv := &http.Server{
		Addr:              listenAddr,
		Handler:           p,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("listening on %s -> %v", listenAddr, upstreams)
	log.Fatal(srv.ListenAndServe())
}
