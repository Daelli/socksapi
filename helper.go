package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"time"
)

func P(path string) string {
	return filepath.Join("data", path)
}

func GEO(socks5 string) (string, error) {
	proxy, err := url.Parse(socks5)
	if err != nil {
		return "", err
	}

	transport := &http.Transport{
		Proxy:               http.ProxyURL(proxy),
		TLSHandshakeTimeout: 5 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	defer transport.CloseIdleConnections()

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	resp, err := client.Get("http://ip-api.com/json")
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	var result struct {
		CountryCode string `json:"countryCode"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.CountryCode == "" {
		return "", fmt.Errorf("CountryCode empty")
	}

	return result.CountryCode, nil
}
