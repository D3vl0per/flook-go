package parser

import (
	"net/http"
	"time"
)

var Client *http.Client

func init() {
	Client = &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: false,
		},
	}
}
