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
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
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

func setupLogging(loggingLevel, logFile string) {
	level := log.DEBUG
	switch loggingLevel {
	case "info":
		level = log.INFO
	case "warn":
		level = log.WARNING
	case "error":
		level = log.ERROR
	}

	log.Global = make(map[string]*log.Filter)

	facility, ok := GetSysLogFacility(logFile)
	if ok {
		flw, err := NewSysLogWriter(facility)
		if err != nil {
			fmt.Fprintf(os.Stderr, "NewSysLogWriter: %s\n", err.Error())
			return
		}
		log.AddFilter("syslog", level, flw)
	} else if logFile == "stdout" {
		flw := log.NewConsoleLogWriter()
		log.AddFilter("stdout", level, flw)
	} else {
		logFileDir := filepath.Dir(logFile)
		os.MkdirAll(logFileDir, 0744)

		flw := log.NewFileLogWriter(logFile, false)
		log.AddFilter("file", level, flw)

		flw.SetFormat("[%D %T] [%L] (%S) %M")
		flw.SetRotate(true)
		flw.SetRotateSize(0)
		flw.SetRotateLines(0)
		flw.SetRotateDaily(true)
	}

	log.Info("Redirectoring logging to %s", logFile)
}

func main() {
	fileName := flag.String("config", "config.toml", "Config file")
	flag.StringVar(&connection, "connection", "", "Connection string to use for Database")
	flag.StringVar(&nsBind, "nameserver_bind", "", "Addr to listen at")
	flag.IntVar(&nsPort, "nameserver_port", 0, "Addr to listen at")
	flag.StringVar(&apiBind, "api_bind", "", "Addr to listen at")
	flag.IntVar(&apiPort, "api_port", 0, "Addr to listen at")
	flag.StringVar(&tsig, "tsig", "", "use MD5 hmac tsig: keyname:base64")
	stdout := flag.Bool("stdout", false, "Log to stdout overriding the configuration")
	syslog := flag.String("syslog", "", "Log to syslog facility overriding the configuration")

	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()

	cfg, err := config.LoadConfiguration(*fileName)

	if err != nil {
		return
	}

	if nsBind != "" {
		cfg.NameServerBind = nsBind
	}
	if nsPort != 0 {
		cfg.NameServerPort = nsPort
	}

	if apiBind != "" {
		cfg.ApiServerBind = apiBind
	}
	if apiPort != 0 {
		cfg.ApiServerPort = apiPort
	}

	if connection != "" {
		cfg.StorageDSN = connection
	}

	if *stdout {
		cfg.LogFile = "stdout"
	}

	if *syslog != "" {
		cfg.LogFile = *syslog
	}

	setupLogging(cfg.LogLevel, cfg.LogFile)

	log.Info("Database is at connection %s", cfg.StorageDSN)

	stats.Setup(cfg)

	db.Setup(cfg)
	// Setup db access
	if db.CheckDB(cfg.StorageDSN) != true {
		log.Warn("Error verifying database connectivity, see above for errors")
		os.Exit(1)
	}

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
