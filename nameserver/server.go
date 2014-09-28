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
package nameserver

import (
	"strings"

	log "code.google.com/p/log4go"
	"github.com/ekarlso/gomdns/config"
	"github.com/miekg/dns"
)

type NameServer struct {
	config  config.Configuration
	stopped bool
}

func NewServer(cfg *config.Configuration) *NameServer {
	s := &NameServer{}
	s.config = *cfg
	return s
}

func (s *NameServer) ListenAndServe() {
	var name, secret string

	if s.config.NameServerSecret != "" {
		a := strings.SplitN(s.config.NameServerSecret, ":", 2)
		name, secret = dns.Fqdn(a[0]), a[1] // fqdn the name, which everybody forgets...
	}

	dns.HandleFunc(".", Handler)
	go s.Serve("tcp", s.config.NameServerBindString(), name, secret)
	go s.Serve("udp", s.config.NameServerBindString(), name, secret)

	s.stopped = false
}

func (s *NameServer) Serve(net, addr, name, secret string) {
	log.Info("Starting NameServer on %s - %s", net, addr)

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

func (s *NameServer) Stop() {
	s.stopped = true
}

/* registers a handlers at the root


// Serve on udp / tcp
go server.Serve("udp", addr, name, secret)
go server.Serve("tcp", addr, name, secret)
*/
