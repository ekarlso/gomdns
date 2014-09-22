package main

import (
	"flag"
	config "github.com/ekarlso/gomdns/config"
	db "github.com/ekarlso/gomdns/db"
	server "github.com/ekarlso/gomdns/server"
	"github.com/miekg/dns"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var (
	connection string
	addr       string
	tsig       string
	bind       string
)

func main() {
	fileName := flag.String("config", "config.sample.toml", "Config file")
	flag.StringVar(&connection, "connection", "designate:designate@tcp(localhost:3306)/designate", "")
	flag.StringVar(&addr, "addr", ":5053", "Addr to listen at")
	flag.StringVar(&tsig, "tsig", "", "use MD5 hmac tsig: keyname:base64")

	var name, secret string
	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()

	config, err := config.LoadConfiguration(*fileName)

	if err != nil {
		return
	}

	log.Printf("Connection %s", config.StorageDSN)
	if tsig != "" {
		a := strings.SplitN(tsig, ":", 2)
		name, secret = dns.Fqdn(a[0]), a[1] // fqdn the name, which everybody forgets...
	}

	// Setup db access
	db.CheckDB(connection)

	// registers a handlers at the root
	dns.HandleFunc(".", server.HandleQuery)

	// Serve on udp / tcp
	go server.Serve("udp", addr, name, secret)
	go server.Serve("tcp", addr, name, secret)

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

forever:
	for {
		select {
		case s := <-sig:
			log.Printf("Signal (%d) received, stopping\n", s)
			break forever
		}
	}
}
