package server

import (
	log "code.google.com/p/log4go"
	"github.com/miekg/dns"
)

func Serve(net, addr, name, secret string) {
	log.Info("Starting server on %s - %s\n", net, addr)

	switch name {
	case "":
		server := &dns.Server{Addr: addr, Net: net, TsigSecret: nil}
		err := server.ListenAndServe()
		if err != nil {
			log.Crash("Failed to setup the "+net+" server: %s\n", err.Error())
		}
	default:
		server := &dns.Server{Addr: addr, Net: net, TsigSecret: map[string]string{name: secret}}
		err := server.ListenAndServe()
		if err != nil {
			log.Crash("Failed to setup the "+net+" server: %s\n", err.Error())
		}
	}
}
