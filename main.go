package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	DATA_API = "https://cdn.jsdelivr.net/gh/proxifly/free-proxy-list@main/proxies/protocols/socks5/data.txt"
	GEO_API  = "http://ip-api.com/json"
	BASE_DIR = "data"
)

var (
	RAW_DATA = bytes.NewBuffer(nil)
	LIMITER  = make(chan struct{}, 10)
	WG       sync.WaitGroup
)

func P(path string) string {
	return filepath.Join(BASE_DIR, path)
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
	resp, err := client.Get(GEO_API)
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

func init() {
	if err := os.RemoveAll(P("")); err != nil {
		log.Fatalln(err)
	}

	if err := os.MkdirAll(P(""), 0755); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	func() {
		log.Println("fetch raw data")

		resp, err := http.Get(DATA_API)
		if err != nil {
			log.Fatalln(err)
		}

		defer resp.Body.Close()
		f, err := os.Create(P("raw.txt"))
		if err != nil {
			log.Fatalln(err)
		}

		defer f.Close()
		io.Copy(io.MultiWriter(f, RAW_DATA), resp.Body)
	}()

	canuse, err := os.Create(P("canuse.txt"))
	if err != nil {
		log.Fatalln(err)
	}

	s := bufio.NewScanner(RAW_DATA)
	for s.Scan() {
		LIMITER <- struct{}{}

		WG.Go(func(proxy string) func() {
			log.Println("process", proxy)

			return func() {
				defer func() { <-LIMITER }()

				_, err := GEO(S5(proxy))
				if err != nil {
					return
				}

				canuse.WriteString(proxy + "\n")
			}
		}(s.Text()))
	}

	defer canuse.Close()
	WG.Wait()
}
