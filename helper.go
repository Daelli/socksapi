package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"time"
)

func P(path string) string {
	return filepath.Join("data", path)
}

func S5(socks5 string) *http.Client {
	proxy, err := url.Parse(socks5)
	if err != nil {
		log.Fatalln(err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxy),

		TLSHandshakeTimeout: 5 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	return &http.Client{
		Transport: transport,
	}
}

func GEO(client *http.Client) (string, error) {
	resp, err := client.Get("http://ip-api.com/json")
	if err != nil {
		return "", err
	}

	var result struct {
		CountryCode string `json:"countryCode"`
	}

	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.CountryCode == "" {
		return "", fmt.Errorf("CountryCode empty")
	}

	return result.CountryCode, nil
}
