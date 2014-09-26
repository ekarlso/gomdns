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
package db

import (
	log "code.google.com/p/log4go"
	config "github.com/ekarlso/gomdns/config"
)

func GetZoneById(zoneId string) (z Zone, err error) {
	cfg := config.GetConfig()
	conn, _ := Connect(cfg.StorageDSN)

	defer conn.Close()

	err = conn.Get(&z, "SELECT id, version, name, email, ttl, serial, refresh, retry, expire, minimum FROM domains WHERE id = ?", zoneId)
	if err != nil {
		log.Debug("Failed getting zone")
	}

	return z, err
}

func GetZoneByName(zoneName string) (z Zone, err error) {
	cfg := config.GetConfig()
	conn, _ := Connect(cfg.StorageDSN)

	defer conn.Close()

	err = conn.Get(&z, "SELECT id, version, name, email, ttl, serial, refresh, retry, expire, minimum FROM domains WHERE name = ?", zoneName)
	if err != nil {
		log.Debug("Failed getting zone")
	}

	return z, err
}

func GetRecordSet(rrName string, rrType string) (rrSet RecordSet, err error) {
	cfg := config.GetConfig()
	conn, _ := Connect(cfg.StorageDSN)

	defer conn.Close()

	err = conn.Get(&rrSet, "SELECT id, domain_id, name, type, ttl from recordsets WHERE name = ? AND type = ?", rrName, rrType)
	if err != nil {
		log.Debug("Failed getting RRSet")
		return rrSet, err
	}

	records := []*Record{}

	err = conn.Select(&records, "SELECT id, domain_id, recordset_id, data, priority, hash FROM records WHERE recordset_id = ?", rrSet.Id)
	if err != nil {
		log.Error("Error fetching records for RRset %s", err)
	}
	rrSet.Records = records
	return rrSet, err
}
