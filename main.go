package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const (
	DATA_API = "https://cdn.jsdelivr.net/gh/proxifly/free-proxy-list@main/proxies/protocols/socks5/data.txt"
	GEO_API  = "http://ip-api.com/line"
	BASE_DIR = "data"
)

func P(path string) string {
	return filepath.Join(BASE_DIR, path)
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
		io.Copy(f, resp.Body)
	}()
}
