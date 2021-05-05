package util

import (
	"io"
	"net"
	"net/http"
	"time"
)

var (
	client = http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
)

// Post wrap http client with timeout
func Post(url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}

// Get wrap http client with timeout
func Get(url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, body)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}
