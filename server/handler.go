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
	"net"
	"strings"
)

func HandleQuery(writer dns.ResponseWriter, req *dns.Msg) {
	cfg := config.GetConfig()

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

	// Deferred write
	defer func() {
		err := writer.WriteMsg(m)
		if err != nil {
			log.Trace(err)
		}
	}()

	// Log the reply
	if cfg.LogQuery == true {
		log.Debug("Query: %v\n", m.String())
	}

	//name := strings.ToLower(query.Name)

	if query.Qtype == dns.TypeSOA {
		soa, _ := SOARecord(query)

		m.Answer = []dns.RR{soa}
		log.Info(m.Answer)
		return
	}

	m.Answer, _ = HandleRRSet(query)
}

func getTtl(rrSet db.RecordSet) (ttl uint32, err error) {
	var (
		zone db.Zone
	)

	zone, err = db.GetZoneById(rrSet.DomainId)

	if err != nil {
		return ttl, err
	}

	if rrSet.Ttl.Valid {
		ttl = uint32(rrSet.Ttl.Int64)
	} else {
		ttl = zone.Ttl
	}
	return ttl, err
}

func HandleRRSet(query dns.Question) (records []dns.RR, err error) {
	log.Info("Attempting to resolve RRSet")

	name := strings.ToLower(query.Name)

	// Attempt to resolve a RRSet and it's Records
	var (
		rrSet  db.RecordSet
		header dns.RR_Header
		ttl    uint32
	)

	rrSet, err = db.GetRecordSet(name, dns.TypeToString[query.Qtype])
	if err != nil {
		log.Error("RecordSet not found", err)
		return records, err
	}

	ttl, err = getTtl(rrSet)

	rrType := dns.StringToType[rrSet.Type]

	header = dns.RR_Header{Name: rrSet.Name, Rrtype: rrType, Class: dns.ClassINET, Ttl: ttl}
	for _, s := range rrSet.Records {
		var record dns.RR

		switch query.Qtype {
		case dns.TypeA:
			log.Debug("IS A ", s)
			record = &dns.A{Hdr: header, A: net.ParseIP(s.Data)}
		}

		log.Debug("Adding record")

		records = append(records, record)
	}
	//records = append(records, &dns.A{Hdr: header, A: ip.To()})

	return records, err
}

func SOARecord(query dns.Question) (soa dns.RR, err error) {
	name := strings.ToLower(query.Name)

	var (
		ttl uint32
	)

	zone, err := db.GetZoneByName(name)
	rrSet, err := db.GetRecordSet(name, "SOA")

	if err != nil {
		return nil, err
	}

	header := dns.RR_Header{Name: zone.Name, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: zone.Ttl}

	// Ttl can be stored on the rrset.Ttl or default to zone.Ttl
	ttl, err = getTtl(rrSet)
	soa = &dns.SOA{
		Hdr:     header,
		Ns:      rrSet.Name,
		Mbox:    strings.Replace(zone.Email, "@", ".", -1) + ".",
		Serial:  uint32(zone.Serial),
		Refresh: uint32(zone.Refresh),
		Retry:   uint32(zone.Retry),
		Expire:  uint32(zone.Expire),
		Minttl:  ttl,
	}

	return soa, err
}
