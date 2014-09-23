/*
 * Copyright (c) 2014 Hewlett-Packard Development Company, L.P.
 *
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package main

import (
	log "code.google.com/p/log4go"
	"flag"
	config "github.com/ekarlso/gomdns/config"
	db "github.com/ekarlso/gomdns/db"
	server "github.com/ekarlso/gomdns/server"
	"github.com/miekg/dns"
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

	log.Info("Database is at connection %s", config.StorageDSN)
	if tsig != "" {
		a := strings.SplitN(tsig, ":", 2)
		name, secret = dns.Fqdn(a[0]), a[1] // fqdn the name, which everybody forgets...
	}

	config.Bind = addr
	config.StorageDSN = connection

	// Setup db access
	db.CheckDB(config.StorageDSN)

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
			log.Info("Signal (%d) received, stopping\n", s)
			break forever
		}
	}
}
