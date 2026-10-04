package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
)

var RAW_DATA = bytes.NewBuffer(nil)

func main() {
	datamgr := NewDataMgr()
	defer datamgr.Close()

	func() {
		log.Println("fetch raw data")

		resp, err := http.Get("https://cdn.jsdelivr.net/gh/proxifly/free-proxy-list@main/proxies/protocols/socks5/data.txt")
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

	limiter := NewLimiter()
	defer limiter.Wait()

	s := bufio.NewScanner(RAW_DATA)
	for s.Scan() {
		limiter.Run(func(proxy string) func() {
			log.Println("process", proxy)

			return func() {
				country, err := GEO(proxy)
				if err != nil {
					return
				}

				datamgr.Add(country, proxy)
			}
		}(s.Text()))
	}
}
