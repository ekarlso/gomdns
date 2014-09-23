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
package server

import (
	log "code.google.com/p/log4go"
	"github.com/miekg/dns"
)

func Serve(net, addr, name, secret string) {
	log.Info("Starting server on %s - %s", net, addr)

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
