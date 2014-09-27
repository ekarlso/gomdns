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
	"net"
	"sort"
	"strings"

	log "code.google.com/p/log4go"
	"github.com/ekarlso/gomdns/config"
	"github.com/ekarlso/gomdns/db"
	"github.com/ekarlso/gomdns/stats"
	"github.com/miekg/dns"
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
			return
		}
	}()

	// Log the reply
	if cfg.LogQuery == true {
		log.Debug("Query: %v\n", m.String())
	}

	stats.AddToMeter("query.total", 1)

	if query.Qtype == dns.TypeSOA {
		soa, _ := SOARecord(query)

		m.Answer = []dns.RR{soa}
		log.Info(m.Answer)
		return
	}

	records, err := ResolveQuery(query)

	if err != nil {
		log.Error("Something went bad: %s", err)
		return
	}

	m.Answer = records
}

// Extract Ttl either from a RRset or a Zone
func resolveRRSetTtl(rrSet db.RecordSet) (ttl uint32, err error) {
	var zone db.Zone

	if rrSet.Ttl.Valid {
		ttl = uint32(rrSet.Ttl.Int64)
	} else {
		log.Debug("Using TTL from domain %s", rrSet.DomainId)

		zone, err = db.GetZoneById(rrSet.DomainId)
		if err != nil {
			return ttl, err
		}

		ttl = zone.Ttl
	}
	return ttl, err
}

// Handle a RRSet
func ResolveQuery(query dns.Question) (records []dns.RR, err error) {
	log.Info("Attempting to resolve RRSet")

	// Attempt to resolve a RRSet and it's Records
	var (
		queryName string
		rrType    string
		rrSet     db.RecordSet
	)

	rrType = dns.TypeToString[query.Qtype]
	queryName = strings.ToLower(query.Name)

	stats.AddToMeter("query."+strings.ToLower(rrType), 1)

	rrSet, err = db.GetRecordSet(queryName, rrType)
	if err != nil {
		log.Error("RecordSet not found", err)
		return records, err
	}

	// Sort MX / SRV by priority
	if query.Qtype == dns.TypeMX || query.Qtype == dns.TypeSRV {
		sort.Sort(db.ByPriority{rrSet.Records})
	}

	records, err = resolveRRSet(query, rrSet)
	return records, err
}

func resolveRRSetHeader(rrSet db.RecordSet) (header dns.RR_Header, err error) {
	var (
		ttl uint32
	)

	ttl, err = resolveRRSetTtl(rrSet)
	if err != nil {
		return header, err
	}

	rrType := dns.StringToType[rrSet.Type]

	header = dns.RR_Header{Name: rrSet.Name, Rrtype: rrType, Class: dns.ClassINET, Ttl: ttl}
	return header, err
}

func resolveRRSet(query dns.Question, rrSet db.RecordSet) (records []dns.RR, err error) {
	header, err := resolveRRSetHeader(rrSet)

	for _, r := range rrSet.Records {
		var record dns.RR

		switch query.Qtype {
		case dns.TypeA:
			record = &dns.A{Hdr: header, A: net.ParseIP(r.Data)}
		case dns.TypeAAAA:
			record = &dns.AAAA{Hdr: header, AAAA: net.ParseIP(r.Data)}
		case dns.TypeCNAME:
			record = &dns.CNAME{Hdr: header, Target: r.Data}
		case dns.TypeMX:
			record = &dns.MX{
				Hdr:        header,
				Preference: uint16(r.Priority.Int64),
				Mx:         r.Data}
		case dns.TypeNS:
			record = &dns.NS{Hdr: header, Ns: r.Data}
		case dns.TypeSRV:
			weight, port, target := extractSrv(r.Data)

			record = &dns.SRV{
				Hdr:      header,
				Priority: uint16(r.Priority.Int64),
				Weight:   weight,
				Port:     port,
				Target:   target,
			}
		case dns.TypeTXT:
			var txt []string
			txt = append(txt, r.Data)

			record = &dns.TXT{
				Hdr: header,
				Txt: txt,
			}
		}

		if record != nil {
			records = append(records, record)
		} else {
			log.Error("Unhandled record")
		}
	}
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

	header := dns.RR_Header{
		Name:   zone.Name,
		Rrtype: dns.TypeSOA,
		Class:  dns.ClassINET,
		Ttl:    zone.Ttl}

	// Ttl can be stored on the rrset.Ttl or default to zone.Ttl
	ttl, err = resolveRRSetTtl(rrSet)

	soa = &dns.SOA{
		Hdr:     header,
		Ns:      rrSet.Name,
		Mbox:    formatEmail(zone.Email),
		Serial:  uint32(zone.Serial),
		Refresh: uint32(zone.Refresh),
		Retry:   uint32(zone.Retry),
		Expire:  uint32(zone.Expire),
		Minttl:  ttl,
	}

	return soa, err
}
