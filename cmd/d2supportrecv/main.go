// Command d2supportrecv receives support bundles from the watch screen's "Send support bundle" button
// (docs/support-bundle.md, cmd/d2supportrecv/README.md). Stdlib only. It listens on a private address behind the
// updates site's web server, never serves the inbox, and logs no address, token or code.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
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
	flag.BoolVar(&c.trustProxy, "trust-proxy", false, "client address from CF-Connecting-IP (set by the tunnel); a request without it is refused. Only behind your own proxy")
	flag.IntVar(&c.globalRate, "global-rate", 200, "all requests per hour together (0: off)")
	flag.IntVar(&c.maxBusy, "max-busy", 3, "uploads handled at once")
	minAge := flag.Int("min-age-hours", 24, "never purge a bundle younger than this; a full inbox then refuses uploads")
	secretFile := flag.String("proxy-secret-file", "", "with -trust-proxy: file holding the secret nginx sends in X-Support-Proxy (CF-Connecting-IP is believed only with it)")
	hash := flag.Bool("hash", false, "read one token from stdin and print its token-file line, then exit")
	flag.Parse()
	if *hash {
		tok, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		tok = strings.TrimSpace(tok)
		if len(tok) < 24 {
			fmt.Fprintln(os.Stderr, "token too short (use at least 24 random characters)")
			os.Exit(2)
		}
		label := "pack"
		if flag.NArg() > 0 {
			label = flag.Arg(0)
		}
		fmt.Println(hashToken(tok), label)
		return
	}
	if c.dir == "" || c.tokensFile == "" {
		fmt.Fprintln(os.Stderr, "-dir and -tokens are required")
		os.Exit(2)
	}
	if c.trustProxy {
		b, err := os.ReadFile(*secretFile)
		c.proxySecret = strings.TrimSpace(string(b))
		if err != nil || len(c.proxySecret) < 16 {
			fmt.Fprintln(os.Stderr, "-trust-proxy needs -proxy-secret-file with at least 16 characters")
			os.Exit(2)
		}
	}
	c.quota <<= 20
	c.minAge = time.Duration(*minAge) * time.Hour
	srv := &http.Server{Addr: *listen, Handler: newServer(c), ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 2 * time.Minute, WriteTimeout: time.Minute, MaxHeaderBytes: 16 << 10, ErrorLog: log.New(io.Discard, "", 0)} // net/http would log the address of a request that panics
	log.Printf("d2supportrecv listening on %s", *listen)
	log.Fatal(srv.ListenAndServe())
}
