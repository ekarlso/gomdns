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

func Handler(writer dns.ResponseWriter, request *dns.Msg) {
	var err error

	query := request.Question[0]

	log.Info("Received query for %s type %s from %s", query.Name, query.Qtype, writer.RemoteAddr())

	stats.AddToMeter("query.total", 1)
	stats.AddToMeter("query."+strings.ToLower(dns.TypeToString[query.Qtype]), 1)

	if query.Qtype == dns.TypeAXFR || query.Qtype == dns.TypeIXFR {
		ResolveXFR(query, writer, request)
		return
	}
	err = ResolveQuery(writer, request)

	if err != nil {
		log.Error("Something went bad: %s", err)
		return
	}
}

func ResolveQuery(writer dns.ResponseWriter, request *dns.Msg) (err error) {
	cfg := config.GetConfig()

	query := request.Question[0]

	m := new(dns.Msg)
	m.SetReply(request)

	// Compress or not
	m.Compress = cfg.CompressQuery
	// We are authoritative..
	m.Authoritative = true

	m.Answer = make([]dns.RR, 0, 10)

	if cfg.LogQuery == true {
		log.Debug("Query: %v\n", m.String())
	}

	// Deferred write
	defer func() {
		err := writer.WriteMsg(m)

		if err != nil {
			log.Trace(err)
			return
		}
	}()

	records, err := ResolveRRSetQuery(query)

	m.Answer = records
	return err
}

func ResolveXFR(query dns.Question, writer dns.ResponseWriter, request *dns.Msg) (err error) {
	// Handle an A|I XFR

	channel := make(chan *dns.Envelope)
	transfer := new(dns.Transfer)
	defer close(channel)

	err = transfer.Out(writer, request, channel)
	if err != nil {
		return
	}

	var (
		records []dns.RR
		rrSets  []db.RecordSet
		zone    db.Zone
	)

	zone, err = db.GetZoneByName(strings.ToLower(query.Name))

	q := dns.Question{Qtype: dns.TypeSOA, Name: query.Name}
	soa, err := ResolveRRSetQuery(q)

	if err != nil {
		log.Error("Error getting SOA for XFR")
		return nil
	}

	rrSets, err = db.GetZoneRecordSets(zone, "", "SOA")
	if err != nil {
		log.Debug("Error getting rrSets and records")
		return err
	}

	records = append(records, soa[0])

	for i, _ := range rrSets {
		rrSetRR, err := resolveRRSet(query, rrSets[i])

		if err != nil {
			log.Error("Error getting RRs for %v, error %v.", query.Name, err)
			return err
		}

		for j, _ := range rrSetRR {
			records = append(records, rrSetRR[j])
		}
	}

	records = append(records, soa[0])

	log.Debug("Records %v", len(records))

	envelope := &dns.Envelope{RR: records}
	channel <- envelope
	writer.Hijack()

	return err
}

// Handle a RRSet
func ResolveRRSetQuery(query dns.Question) (records []dns.RR, err error) {
	log.Info("Attempting to resolve RRSet")

	// Attempt to resolve a RRSet and it's Records
	var (
		queryName string
		rrType    string
		rrSet     db.RecordSet
	)

	rrType = dns.TypeToString[query.Qtype]
	queryName = strings.ToLower(query.Name)

	rrSet, err = db.GetRecordSet(queryName, rrType)
	if err != nil {
		log.Error("RecordSet not found", err)
		return records, err
	}

	records, err = resolveRRSet(query, rrSet)
	return records, err
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

// Create a header
func resolveRRSetHeader(rrSet db.RecordSet) (header dns.RR_Header, err error) {
	var ttl uint32

	ttl, err = resolveRRSetTtl(rrSet)
	if err != nil {
		return header, err
	}
	rrType := dns.StringToType[rrSet.Type]

	header = dns.RR_Header{Name: rrSet.Name, Rrtype: rrType, Class: dns.ClassINET, Ttl: ttl}
	return header, err
}

// Create dns.RR records from a query and rrSet
func resolveRRSet(query dns.Question, rrSet db.RecordSet) (records []dns.RR, err error) {
	if len(rrSet.Records) == 0 {
		log.Debug("No records on RRSet %v", rrSet.Id)
		return records, err
	}

	rrType := dns.StringToType[rrSet.Type]

	// Sort MX / SRV by priority
	if rrType == dns.TypeMX || rrType == dns.TypeSRV {
		sort.Sort(db.ByPriority{rrSet.Records})
	}

	header, err := resolveRRSetHeader(rrSet)

	for _, r := range rrSet.Records {
		var record dns.RR

		switch rrType {
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
		case dns.TypeSOA:
			soa := r.ExtractSOA()
			record = &dns.SOA{
				Hdr:     header,
				Ns:      soa.Ns,
				Mbox:    soa.Mbox,
				Serial:  soa.Serial,
				Refresh: soa.Refresh,
				Retry:   soa.Retry,
				Expire:  soa.Expire,
				Minttl:  soa.Minttl,
			}

		case dns.TypeSRV:
			srv := r.ExtractSrv()

			record = &dns.SRV{
				Hdr:      header,
				Priority: uint16(r.Priority.Int64),
				Weight:   srv.Weight,
				Port:     srv.Port,
				Target:   srv.Target,
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
