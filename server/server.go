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
	"github.com/ekarlso/gomdns/api"
	"github.com/ekarlso/gomdns/config"
	"github.com/ekarlso/gomdns/nameserver"
)

type Server struct {
	ApiServer  *api.HttpServer
	NameServer *nameserver.NameServer
	Config     *config.Configuration
	stopped    bool
}

func NewServer(cfg *config.Configuration) (*Server, error) {
	apiServer := api.NewServer(cfg)
	nameServer := nameserver.NewServer(cfg)

	return &Server{
		ApiServer:  apiServer,
		NameServer: nameServer,
		Config:     cfg,
	}, nil
}

func (self *Server) ListenAndServe() (err error) {
	log.Debug("Starting API on %s", self.Config.ApiServerListen())
	go self.ApiServer.ListenAndServe()

	self.NameServer.ListenAndServe()

	return err
}

func (self *Server) Stop() {
	if self.stopped {
		return
	}

	self.stopped = true
}
