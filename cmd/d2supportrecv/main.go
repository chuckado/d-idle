// Command d2supportrecv receives support bundles from the watch screen's "Send support bundle" button
// (docs/support-bundle.md, cmd/d2supportrecv/README.md). Stdlib only. It listens on a private address behind the
// updates site's web server, never serves the inbox, and logs no address, token or code.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	var c config
	listen := flag.String("listen", "127.0.0.1:8899", "address to listen on (keep it private; the web server proxies to it)")
	flag.StringVar(&c.dir, "dir", "", "inbox directory (never web-served)")
	flag.StringVar(&c.tokensFile, "tokens", "", "token file: one line per upload token, `<sha256 hex> [label]`")
	flag.Int64Var(&c.quota, "quota-mb", 2048, "inbox quota in MB; the oldest bundles are deleted first")
	flag.IntVar(&c.maxFiles, "max-files", 500, "inbox file limit")
	flag.IntVar(&c.tokenRate, "token-rate", 20, "uploads per hour per token")
	flag.IntVar(&c.ipRate, "ip-rate", 10, "requests per hour per client address")
	flag.IntVar(&c.failRate, "fail-rate", 5, "failed authentications per hour per client address")
	flag.BoolVar(&c.trustProxy, "trust-proxy", false, "take the client address from CF-Connecting-IP / X-Forwarded-For (only behind your own proxy)")
	hash := flag.Bool("hash", false, "read one token from stdin and print its token-file line, then exit")
	flag.Parse()
	if *hash {
		tok, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		tok = strings.TrimSpace(tok)
		if len(tok) < 24 {
			fmt.Fprintln(os.Stderr, "token too short (use at least 24 random characters)")
			os.Exit(2)
		}
		fmt.Println(hashToken(tok), "pack")
		return
	}
	if c.dir == "" || c.tokensFile == "" {
		fmt.Fprintln(os.Stderr, "-dir and -tokens are required")
		os.Exit(2)
	}
	c.quota <<= 20
	srv := &http.Server{Addr: *listen, Handler: newServer(c), ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 2 * time.Minute, WriteTimeout: time.Minute, MaxHeaderBytes: 16 << 10}
	log.Printf("d2supportrecv listening on %s", *listen)
	log.Fatal(srv.ListenAndServe())
}
