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
	"flag"
	"os"
	"os/signal"
	"syscall"

	log "code.google.com/p/log4go"
	"github.com/ekarlso/gomdns/config"
	"github.com/ekarlso/gomdns/db"
	"github.com/ekarlso/gomdns/server"
	"github.com/ekarlso/gomdns/stats"
)

var (
	connection string
	nsBind     string
	nsPort     int
	apiBind    string
	apiPort    int

	tsig string
)

func main() {
	fileName := flag.String("config", "config.sample.toml", "Config file")
	flag.StringVar(&connection, "connection", "designate:designate@tcp(localhost:3306)/designate", "Connection string to use for Database")
	flag.StringVar(&nsBind, "nameserver_bind", "", "Addr to listen at")
	flag.IntVar(&nsPort, "nameserver_port", 5053, "Addr to listen at")
	flag.StringVar(&apiBind, "api_bind", "", "Addr to listen at")
	flag.IntVar(&apiPort, "api_port", 5080, "Addr to listen at")
	flag.StringVar(&tsig, "tsig", "", "use MD5 hmac tsig: keyname:base64")

	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()

	cfg, err := config.LoadConfiguration(*fileName)

	if err != nil {
		return
	}

	cfg.NameServerBind = nsBind
	cfg.NameServerPort = nsPort
	cfg.ApiServerBind = apiBind
	cfg.ApiServerPort = apiPort
	cfg.StorageDSN = connection

	log.Info("Database is at connection %s", cfg.StorageDSN)

	// Setup db access
	if db.CheckDB(cfg.StorageDSN) != true {
		log.Warn("Error verifying database connectivity, see above for errors")
		os.Exit(1)
	}

	stats.Setup()

	srv, err := server.NewServer(cfg)
	srv.ListenAndServe()

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
