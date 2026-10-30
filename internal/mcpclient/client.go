package mcpclient

import (
	"net/http"
	"time"
)

type Client struct {
	Timeout    time.Duration
	HTTPClient *http.Client
}

func New() *Client {
	return &Client{Timeout: 10 * time.Second}
}

type bearerTransport struct {
	base http.RoundTripper
	key  string
}

func (t bearerTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	clone := request.Clone(request.Context())
	clone.Header.Set("Authorization", "Bearer "+t.key)
	return t.base.RoundTrip(clone)
}

func (c *Client) authenticatedHTTPClient(key string) *http.Client {
	var client http.Client
	if c.HTTPClient != nil {
		client = *c.HTTPClient
	}
	base := client.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	client.Transport = bearerTransport{base: base, key: key}
	return &client
}
