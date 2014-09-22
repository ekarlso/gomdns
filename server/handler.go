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
	"github.com/ekarlso/gomdns/config"
	"github.com/ekarlso/gomdns/db"
	"github.com/miekg/dns"
	"strings"
)

func HandleQuery(writer dns.ResponseWriter, req *dns.Msg) {
	cfg := config.GetConfig()

	var (
		zone db.Zone
	)

	query := req.Question[0]

	log.Info("Received query for %s type %s from %s", query.Name, query.Qtype, writer.RemoteAddr())

	// Create a new msg and set the reply
	m := new(dns.Msg)
	m.SetReply(req)

	// Compress or not
	m.Compress = cfg.CompressQuery
	// We are authoritative..
	m.Authoritative = true

	m.Answer = make([]dns.RR, 0, 10)

	defer func() {
		err := writer.WriteMsg(m)
		if err != nil {
			log.Trace(err)
		}
	}()

	// Check of we have the zone in our db.
	zoneName := strings.ToLower(query.Name)
	if _, ok := dns.IsDomainName(zoneName); ok {
		log.Info("Name %s is Zone", zoneName)

		var err error
		err, zone = db.GetZoneByName(zoneName)
		if err != nil {
			log.Warn("Zone %s wasn't found", zoneName)
			return
		}
	} else {
		log.Warn("Request %s is not a valid Zone")
		return
	}

	// Log the reply
	if cfg.LogQuery == true {
		log.Debug("Query: %v\n", m.String())
	}

	if query.Qtype == dns.TypeSOA {
		soa, _ := SOARecord(zone)

		m.Answer = []dns.RR{soa}
		log.Info(m.Answer)
		return
	}
}

func SOARecord(zone db.Zone) (soa dns.RR, err error) {
	rrset, err := db.GetRecordSet(zone.Id, "SOA")

	if err != nil {
		return nil, err
	}

	header := dns.RR_Header{Name: zone.Name, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: zone.Ttl}

	var ttl uint32
	if rrset.Ttl.Valid {
		ttl = uint32(rrset.Ttl.Int64)
		log.Debug("%v", rrset.Ttl.Int64)
	} else {
		ttl = zone.Ttl
	}

	soa = &dns.SOA{
		Hdr:     header,
		Ns:      rrset.Name,
		Mbox:    strings.Replace(zone.Email, "@", ".", -1) + ".",
		Serial:  uint32(zone.Serial),
		Refresh: uint32(zone.Refresh),
		Retry:   uint32(zone.Retry),
		Expire:  uint32(zone.Expire),
		Minttl:  ttl,
	}

	return soa, err
}
