package server

import (
	"github.com/ekarlso/gomdns/config"
	"github.com/ekarlso/gomdns/db"

	"github.com/miekg/dns"
	"log"
)

func HandleQuery(w dns.ResponseWriter, r *dns.Msg) {
	log.Printf("Handling query")

	_ = db.Connect(config.Configuration.StorageDSN)
}
