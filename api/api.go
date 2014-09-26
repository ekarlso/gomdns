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
package api

import (
	"net"
	libhttp "net/http"
	"strings"
	"time"

	log "code.google.com/p/log4go"
	"github.com/ekarlso/gomdns/config"
	tiger "github.com/rcrowley/go-tigertonic"
)

type HttpServer struct {
	conn        net.Listener
	httpPort    string
	shutdown    chan bool
	config      *config.Configuration
	readTimeout time.Duration
	mux         *tiger.HostServeMux
}

func NewServer(config *config.Configuration) *HttpServer {
	self := &HttpServer{}
	self.httpPort = config.ApiServerListen()
	self.shutdown = make(chan bool, 2)
	self.config = config
	self.mux = &tiger.NewHostServeMux()
	return self
}

func (self *HttpServer) ListenAndServe() (err error) {
	self.conn, err = net.Listen("tcp", self.httpPort)
	return err
}

func (self *HttpServer) Serve(listener net.Listener) {
	defer func() { self.shutdown <- true }()

	self.conn = listener

	self.mux.Handle("GET", "/stats", self.getStats)

	self.serveListener(listener, self.mux)
}

func (self *HttpServer) serveListener(listener net.Listener, m *tiger.HostServeMux) {
	srv := &libhttp.Server{Handler: m, ReadTimeout: self.readTimeout}
	if err := srv.Serve(listener); err != nil && !strings.Contains(err.Error(), "closed network") {
		panic(err)
	}
}

func (self *HttpServer) Close() {
	if self.conn != nil {
		log.Info("Closing http server")
		self.conn.Close()
		log.Info("Waiting for all requests to finish before killing the process")
		select {
		case <-time.After(time.Second * 5):
			log.Error("There seems to be a hanging request. Closing anyway")
		case <-self.shutdown:
		}
	}
}

func (self *HttpServer) getStats() {}
