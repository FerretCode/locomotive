package webhook

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

var client *http.Client

func init() {
	client = &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 5 * time.Minute,
			}).DialContext,
			MaxConnsPerHost:       100,
			MaxIdleConns:          100,
			IdleConnTimeout:       5 * time.Minute,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
			ExpectContinueTimeout: 5 * time.Second,
			MaxIdleConnsPerHost:   100,
			DisableKeepAlives:     false,
			TLSClientConfig:       &tls.Config{},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
