package proxy

import (
	"net/http"
	"net/url"
)

type Balancer interface {
	Pick(r *http.Request) (up *url.URL, done func(), ok bool)
}
