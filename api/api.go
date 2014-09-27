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
	"net/url"
	"strings"
	"time"

	log "code.google.com/p/log4go"
	"github.com/ekarlso/gomdns/config"
	"github.com/ekarlso/gomdns/stats"
	metrics "github.com/rcrowley/go-metrics"
	tiger "github.com/rcrowley/go-tigertonic"
)

type Request struct {
}

type Stats struct {
	Stats []Stat
}
type Stat struct {
	Name string
}

type HttpServer struct {
	conn        net.Listener
	httpPort    string
	shutdown    chan bool
	config      *config.Configuration
	readTimeout time.Duration
	mux         *tiger.TrieServeMux
}

func NewServer(config *config.Configuration) *HttpServer {
	self := &HttpServer{}
	self.httpPort = config.ApiServerListen()
	self.shutdown = make(chan bool, 2)
	self.config = config
	self.mux = tiger.NewTrieServeMux()
	return self
}

func (self *HttpServer) ListenAndServe() {
	var err error
	self.conn, err = net.Listen("tcp", self.httpPort)

	if err != nil {
		log.Error("Listen: ", err)
	}

	self.Serve(self.conn)
}

func (self *HttpServer) Serve(listener net.Listener) {
	defer func() { self.shutdown <- true }()

	self.conn = listener

	self.mux.Handle("GET", "/stats", tiger.Marshaled(self.getStats))

	self.serveListener(listener, self.mux)
}

func (self *HttpServer) serveListener(listener net.Listener, m *tiger.TrieServeMux) {
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

func (self *HttpServer) getStats(u *url.URL, h libhttp.Header, req *Request) (int, libhttp.Header, metrics.Registry, error) {
	return libhttp.StatusOK, nil, stats.NameServerStats, nil
}
