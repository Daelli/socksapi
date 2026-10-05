package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// 多个候选源，任一挂掉不影响整体；全部以 socks5://ip:port 归一化后去重
var SOURCES = []string{
	"https://cdn.jsdelivr.net/gh/proxifly/free-proxy-list@main/proxies/protocols/socks5/data.txt",
	"https://cdn.jsdelivr.net/gh/TheSpeedX/PROXY-List@master/socks5.txt",
	"https://cdn.jsdelivr.net/gh/ALIILAPRO/Proxy@main/socks5.txt",
	"https://cdn.jsdelivr.net/gh/monosans/proxy-list@main/proxies/socks5.txt",
	"https://cdn.jsdelivr.net/gh/Zaeem20/FREE_PROXIES_LIST@master/socks5.txt",
}

var RAW_DATA = bytes.NewBuffer(nil)

// 只保留第一个源的原始清单，免得 raw.txt 被后面的源覆盖
var rawWritten bool

func normalize(line string) string {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return ""
	}
	if !strings.Contains(line, "://") {
		line = "socks5://" + line
	}
	if !strings.HasPrefix(line, "socks5://") {
		return ""
	}
	hostport := strings.TrimPrefix(line, "socks5://")
	if strings.Count(hostport, ":") < 1 {
		return ""
	}
	return "socks5://" + hostport
}

func main() {
	datamgr := NewDataMgr()
	defer datamgr.Close()

	seen := map[string]bool{}
	total := 0

	for _, src := range SOURCES {
		func() {
			log.Println("fetch", src)

			resp, err := http.Get(src)
			if err != nil {
				log.Println("skip source:", err)
				return
			}

			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Println("skip source: http", resp.StatusCode)
				return
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Println("skip source: read", err)
				return
			}

			if !rawWritten {
				rawWritten = true
				f, err := os.Create(P("raw.txt"))
				if err != nil {
					log.Fatalln(err)
				}
				defer f.Close()
				f.Write(body)
			}

			s := bufio.NewScanner(bytes.NewReader(body))
			s.Buffer(make([]byte, 1024*1024), 1024*1024)
			for s.Scan() {
				proxy := normalize(s.Text())
				if proxy == "" || seen[proxy] {
					continue
				}
				seen[proxy] = true
				total++
				RAW_DATA.WriteString(proxy + "\n")
			}
		}()
	}

	log.Println("candidates:", total)

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
